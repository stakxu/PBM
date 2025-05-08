package core

import (
	"agent/config"
	"agent/protocol"
	"context"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
	"runtime"
	"strings"
	"time"
)

type Collector struct {
	cfg        *config.Config
	stop       chan struct{}
	lastInOut  map[string]net.IOCountersStat
	lastStatic time.Time
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewCollector(cfg *config.Config) *Collector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Collector{
		cfg:       cfg,
		stop:      make(chan struct{}),
		lastInOut: make(map[string]net.IOCountersStat),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (c *Collector) Stop() error {
	c.cancel()
	close(c.stop)
	return nil
}

func (c *Collector) collectStaticInfo() (*protocol.StaticSystemInfo, error) {
	info := &protocol.StaticSystemInfo{
		UUID:     GetAgentUUID(),
		Alias:    c.cfg.Agent.Alias,
		UpdateAt: time.Now().Unix(),
	}

	// 获取 CPU 信息
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.CPU.Model = cpuInfo[0].ModelName // 使用实际的 CPU 型号
	} else {
		info.CPU.Model = "Unknown CPU" // 如果获取失败则使用默认值
	}
	info.CPU.Cores = runtime.NumCPU()

	// 获取内存信息
	if memory, err := memory.Get(); err == nil {
		info.Memory.Total = memory.Total
		info.Swap.Total = memory.SwapTotal
	}

	// 获取磁盘信息
	if parts, err := disk.Partitions(false); err == nil {
		for _, part := range parts {
			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				info.Disk.Total += usage.Total
			}
		}
	}

	// 获取网络接口信息
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			// 跳过回环接口、非活动接口和 Docker 接口
			if strings.Contains(iface.Name, "lo") || 
			   !strings.Contains(strings.Join(iface.Flags, " "), "up") ||
			   strings.Contains(iface.Name, "docker") ||
			   strings.Contains(iface.Name, "veth") {
				continue
			}

			for _, addr := range iface.Addrs {
				ip := addr.Addr
				// 移除 CIDR 后缀
				if idx := strings.Index(ip, "/"); idx != -1 {
					ip = ip[:idx]
				}

				// 跳过私有地址
				if isPrivateIP(ip) {
					continue
				}

				if isIPv4(ip) {
					info.IPv4 = append(info.IPv4, ip)
				} else if isIPv6(ip) && !isLinkLocalIPv6(ip) {
					info.IPv6 = append(info.IPv6, ip)
				}
			}
		}
	}

	return info, nil
}

func (c *Collector) collectDynamicInfo() (*protocol.SystemInfo, error) {
	info := &protocol.SystemInfo{
		UUID: GetAgentUUID(),
	}

	// 获取网络流量
	if netIO, err := net.IOCounters(true); err == nil {
		for _, io := range netIO {
			if last, ok := c.lastInOut[io.Name]; ok {
				info.NetworkTraffic.In += io.BytesRecv - last.BytesRecv
				info.NetworkTraffic.Out += io.BytesSent - last.BytesSent
			}
			c.lastInOut[io.Name] = io
		}
	}

	// 获取系统运行时间
	if uptime, err := host.Uptime(); err == nil {
		info.Uptime = float64(uptime)
	}

	// 获取 CPU 使用率
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		info.CPU.Usage = cpuPercent[0]
	}

	// 获取内存使用情况
	if memory, err := memory.Get(); err == nil {
		info.Memory.Used = memory.Used
		info.Swap.Used = memory.SwapUsed
	}

	// 获取磁盘使用情况
	if parts, err := disk.Partitions(false); err == nil {
		for _, part := range parts {
			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				info.Disk.Used += usage.Used
			}
		}
	}

	// 获取网络连接数
	if conns, err := net.Connections("all"); err == nil {
		for _, conn := range conns {
			switch conn.Type {
			case 1: // TCP
				info.Network.TCP++
			case 2: // UDP
				info.Network.UDP++
			}
		}
	}

	return info, nil
}

// 判断是否为私有 IP 地址
func isPrivateIP(ip string) bool {
	// 检查 IPv4 私有地址范围
	if strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "172.17.") ||
		strings.HasPrefix(ip, "172.18.") ||
		strings.HasPrefix(ip, "172.19.") ||
		strings.HasPrefix(ip, "172.20.") ||
		strings.HasPrefix(ip, "172.21.") ||
		strings.HasPrefix(ip, "172.22.") ||
		strings.HasPrefix(ip, "172.23.") ||
		strings.HasPrefix(ip, "172.24.") ||
		strings.HasPrefix(ip, "172.25.") ||
		strings.HasPrefix(ip, "172.26.") ||
		strings.HasPrefix(ip, "172.27.") ||
		strings.HasPrefix(ip, "172.28.") ||
		strings.HasPrefix(ip, "172.29.") ||
		strings.HasPrefix(ip, "172.30.") ||
		strings.HasPrefix(ip, "172.31.") ||
		strings.HasPrefix(ip, "192.168.") {
		return true
	}
	return false
}

// 判断是否为 IPv4 地址
func isIPv4(addr string) bool {
	for i := 0; i < len(addr); i++ {
		if addr[i] == '.' {
			return true
		}
	}
	return false
}

// 判断是否为 IPv6 地址
func isIPv6(addr string) bool {
	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' {
			return true
		}
	}
	return false
}

// 判断是否为链路本地 IPv6 地址
func isLinkLocalIPv6(addr string) bool {
	return strings.HasPrefix(addr, "fe80:")
}