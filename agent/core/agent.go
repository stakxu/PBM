package core

import (
	"agent/config"
	"agent/logger"
	"agent/plugin"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"sync"
)

var (
	agentUUID     string
	agentUUIDOnce sync.Once
)

type Agent struct {
	cfg       *config.Config
	client    *Client
	collector *Collector
	plugins   *plugin.Manager
	stopWg    sync.WaitGroup
}

func NewAgent(cfg *config.Config) *Agent {
	collector := NewCollector(cfg)
	client := NewClient(cfg)
	client.SetCollector(collector)

	return &Agent{
		cfg:       cfg,
		client:    client,
		collector: collector,
		plugins:   plugin.NewManager(),
	}
}

func (a *Agent) Start() error {
	// 启动 TCP 客户端
	if err := a.client.Start(); err != nil {
		return err
	}

	logger.Info("Agent 启动成功")
	return nil
}

func (a *Agent) Stop() error {
	logger.Info("正在停止 Agent...")
	
	// 停止所有插件
	a.plugins.StopAll()

	// 停止系统信息采集器
	if err := a.collector.Stop(); err != nil {
		logger.Error("停止系统信息采集器失败:", err)
	}

	// 停止 TCP 客户端并等待所有 goroutine 完成
	if err := a.client.Stop(); err != nil {
		logger.Error("停止 TCP 客户端失败:", err)
	}

	logger.Info("Agent 已停止")
	return nil
}

func GetAgentUUID() string {
	agentUUIDOnce.Do(func() {
		// 尝试从文件读取 UUID
		uuidFile := filepath.Join("data", "agent.uuid")
		if data, err := os.ReadFile(uuidFile); err == nil {
			agentUUID = string(data)
			return
		}

		// 生成新的 UUID
		agentUUID = uuid.New().String()

		// 保存到文件
		os.MkdirAll("data", 0755)
		os.WriteFile(uuidFile, []byte(agentUUID), 0644)
	})
	return agentUUID
}