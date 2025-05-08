package core

import (
	"agent/config"
	"agent/logger"
	"agent/protocol"
	"fmt"
	"net"
	"sync"
	"time"
)

type Client struct {
	cfg         *config.Config
	conn        net.Conn
	parser      *protocol.MessageParser
	connected   bool
	mutex       sync.RWMutex
	reconnect   chan struct{}
	stop        chan struct{}
	systemInfo  chan *protocol.SystemInfo
	staticInfo  chan *protocol.StaticSystemInfo
	heartbeat   *time.Ticker
	registered  bool
	collector   *Collector
	stopWg      sync.WaitGroup
	shutdownOnce sync.Once
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:        cfg,
		parser:     protocol.NewMessageParser(),
		reconnect:  make(chan struct{}, 1), // 使用带缓冲的channel
		stop:       make(chan struct{}),
		systemInfo: make(chan *protocol.SystemInfo, 100),
		staticInfo: make(chan *protocol.StaticSystemInfo, 10),
	}
}

func (c *Client) Start() error {
	logger.Info("开始启动客户端连接...")
	// 启动连接管理
	c.stopWg.Add(1)
	go func() {
		defer c.stopWg.Done()
		c.connectionManager()
	}()
	
	// 启动心跳
	c.stopWg.Add(1)
	go func() {
		defer c.stopWg.Done()
		c.heartbeatManager()
	}()
	
	// 触发首次连接
	c.reconnect <- struct{}{}
	
	return nil
}

func (c *Client) Stop() error {
	c.shutdownOnce.Do(func() {
		logger.Info("正在停止客户端...")
		close(c.stop)
		
		c.mutex.Lock()
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.connected = false
		c.mutex.Unlock()
		
		if c.heartbeat != nil {
			c.heartbeat.Stop()
		}
		
		// 等待所有 goroutine 完成
		c.stopWg.Wait()
		logger.Info("客户端已完全停止")
	})
	return nil
}


func (c *Client) connectionManager() {
	logger.Info("连接管理器启动")
	for {
		select {
		case <-c.stop:
			logger.Info("连接管理器收到停止信号")
			return
		case <-c.reconnect:
			logger.Info("尝试建立连接...")
			c.connect()
		}
	}
}

func (c *Client) connect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		logger.Info("已经连接到服务器，跳过连接")
		return
	}

	// 尝试所有可用地址
	addresses := []string{c.cfg.Hub.Address}
	addresses = append(addresses, c.cfg.Hub.BackupAddresses...)

	for _, addr := range addresses {
		endpoint := fmt.Sprintf("%s:%d", addr, c.cfg.Hub.Port)
		logger.Info("正在连接到服务器:", endpoint)
		
		// 根据配置选择网络协议
		network := "tcp"
		if c.cfg.Hub.Protocol == "ipv6" {
			network = "tcp6"
		} else if c.cfg.Hub.Protocol == "ipv4" {
			network = "tcp4"
		}
		logger.Info("使用网络协议:", network)

		conn, err := net.DialTimeout(network, endpoint, 5*time.Second)
		if err != nil {
			logger.Error("连接失败:", addr, err)
			continue
		}

		logger.Info("成功建立TCP连接")
		c.conn = conn
		c.connected = true
		
		// 发送认证消息
		authMsg := protocol.NewMessage(protocol.MessageTypeAuth, &protocol.AuthPayload{
			Key:   c.cfg.Auth.Key,
			UUID:  GetAgentUUID(),
			Alias: c.cfg.Agent.Alias,
		})
		
		logger.Info("正在发送认证消息...")
		logger.Debug("认证消息内容:", authMsg)
		
		data := authMsg.Encode()
		logger.Debug("编码后的认证消息:", fmt.Sprintf("%x", data))
		
		n, err := c.conn.Write(data)
		if err != nil {
			logger.Error("发送认证消息失败:", err)
			c.conn.Close()
			c.conn = nil
			c.connected = false
			continue
		}
		logger.Info("认证消息发送成功, 已发送", n, "字节")

		// 启动接收循环
		c.stopWg.Add(1)
		go func() {
			defer c.stopWg.Done()
			c.receiveLoop()
		}()

		// 发送静态系统信息
		if staticInfo, err := c.collector.collectStaticInfo(); err == nil {
			staticInfoMsg := protocol.NewMessage(protocol.MessageTypeStaticInfo, staticInfo)
			logger.Debug("静态系统信息内容:", staticInfoMsg)
			data := staticInfoMsg.Encode()
			logger.Debug("编码后的静态系统信息:", fmt.Sprintf("%x", data))
			
			n, err := c.conn.Write(data)
			if err != nil {
				logger.Error("发送静态系统信息失败:", err)
			} else {
				logger.Info("静态系统信息发送成功, 已发送", n, "字节")
			}
		}

		// 启动系统信息定时上报
		c.stopWg.Add(1)
		go func() {
			defer c.stopWg.Done()
			c.systemInfoReporter()
		}()
		return
	}

	// 所有地址都连接失败,等待重试
	retryInterval := time.Duration(c.cfg.Agent.ReconnectInterval) * time.Second
	logger.Info("所有连接尝试失败，将在", retryInterval, "后重试")
	time.Sleep(retryInterval)
	
	// 使用 select 避免阻塞
	select {
	case c.reconnect <- struct{}{}:
	default:
		// channel 已满，说明重连已在进行中
	}
}

func (c *Client) systemInfoReporter() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("系统信息上报器发生panic:", r)
			go c.handleDisconnect()
		}
	}()

	ticker := time.NewTicker(time.Duration(c.cfg.Agent.SystemInfoInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.mutex.RLock()
			if !c.connected || c.conn == nil {
				c.mutex.RUnlock()
				continue
			}
			
			if info, err := c.collector.collectDynamicInfo(); err == nil {
				msg := protocol.NewMessage(protocol.MessageTypeSystemInfo, info)
				logger.Debug("系统信息内容:", msg)
				data := msg.Encode()
				logger.Debug("编码后的系统信息:", fmt.Sprintf("%x", data))
				
				n, err := c.conn.Write(data)
				if err != nil {
					logger.Error("发送系统信息失败:", err)
					c.mutex.RUnlock()
					c.handleDisconnect()
					continue
				}
				logger.Debug("系统信息发送成功, 已发送", n, "字节")
			}
			c.mutex.RUnlock()
		}
	}
}

func (c *Client) receiveLoop() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("接收循环发生panic:", r)
			go c.handleDisconnect()
		}
	}()

	logger.Info("启动数据接收循环")
	buffer := make([]byte, 4096)
	
	for {
		select {
		case <-c.stop:
			return
		default:
			c.mutex.RLock()
			if !c.connected || c.conn == nil {
				c.mutex.RUnlock()
				return
			}
			
			// 设置读取超时
			c.conn.SetReadDeadline(time.Now().Add(time.Second * 30))
			
			n, err := c.conn.Read(buffer)
			if err != nil {
				c.mutex.RUnlock()
				logger.Error("读取数据失败:", err)
				c.handleDisconnect()
				return
			}

			logger.Debug("收到", n, "字节数据:", fmt.Sprintf("%x", buffer[:n]))
			c.parser.Append(buffer[:n])
			
			for c.parser.HasCompleteMessage() {
				if msg := c.parser.ParseMessage(); msg != nil {
					logger.Info("解析到完整消息:", msg.Header.Type)
					logger.Debug("消息内容:", msg)
					c.handleMessage(msg)
				}
			}
			c.mutex.RUnlock()
		}
	}
}

func (c *Client) handleMessage(msg *protocol.Message) {
	switch msg.Header.Type {
	case protocol.MessageTypeConfig:
		logger.Info("收到配置更新消息")
		logger.Debug("配置内容:", msg.Payload)
		// 处理配置更新
		var config struct {
			SystemInfoInterval int `json:"systemInfoInterval"`
			HeartbeatInterval int `json:"heartbeatInterval"`
		}
		if err := msg.DecodePayload(&config); err == nil {
			c.updateIntervals(config.SystemInfoInterval, config.HeartbeatInterval)
		}
	}
}

func (c *Client) handleDisconnect() {
	logger.Info("处理连接断开")
	c.mutex.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.mutex.Unlock()

	// 触发重连
	logger.Info("触发重连机制")
	select {
	case c.reconnect <- struct{}{}:
	default:
		// channel 已满，说明重连已在进行中
	}
}

func (c *Client) heartbeatManager() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("心跳管理器发生panic:", r)
			go c.handleDisconnect()
		}
	}()

	logger.Info("心跳管理器启动")
	c.heartbeat = time.NewTicker(time.Duration(c.cfg.Agent.HeartbeatInterval) * time.Second)
	defer c.heartbeat.Stop()

	for {
		select {
		case <-c.stop:
			logger.Info("心跳管理器收到停止信号")
			return
		case <-c.heartbeat.C:
			c.mutex.RLock()
			if !c.connected || c.conn == nil {
				c.mutex.RUnlock()
				continue
			}
			
			heartbeat := protocol.NewMessage(protocol.MessageTypeHeartbeat, &protocol.HeartbeatPayload{
				UUID: GetAgentUUID(),
			})
			data := heartbeat.Encode()
			logger.Debug("发送心跳:", fmt.Sprintf("%x", data))
			n, err := c.conn.Write(data)
			if err != nil {
				logger.Error("发送心跳失败:", err)
				c.mutex.RUnlock()
				c.handleDisconnect()
				continue
			}
			logger.Debug("心跳发送成功, 已发送", n, "字节")
			c.mutex.RUnlock()
		}
	}
}

func (c *Client) updateIntervals(systemInfo, heartbeat int) {
	if heartbeat > 0 && heartbeat != c.cfg.Agent.HeartbeatInterval {
		logger.Info("更新心跳间隔:", heartbeat, "秒")
		c.cfg.Agent.HeartbeatInterval = heartbeat
		if c.heartbeat != nil {
			c.heartbeat.Reset(time.Duration(heartbeat) * time.Second)
		}
	}
}

func (c *Client) Send(msg *protocol.Message) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected || c.conn == nil {
		return fmt.Errorf("未连接到服务器")
	}

	data := msg.Encode()
	logger.Debug("发送消息:", msg.Header.Type, "大小:", len(data), "字节")
	logger.Debug("消息内容:", fmt.Sprintf("%x", data))
	n, err := c.conn.Write(data)
	if err != nil {
		return err
	}
	logger.Debug("消息发送成功, 已发送", n, "字节")
	return nil
}

func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

func (c *Client) SetCollector(collector *Collector) {
	c.collector = collector
}
