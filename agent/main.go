package main

import (
	"agent/config"
	"agent/core"
	"agent/logger"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 初始化日志
	if err := logger.Init(); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Close()

	logger.Info("Agent 正在启动...")

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("加载配置失败: ", err)
		os.Exit(1)
	}

	// 创建并启动 Agent
	agent := core.NewAgent(cfg)
	if err := agent.Start(); err != nil {
		logger.Error("启动 Agent 失败: ", err)
		os.Exit(1)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭
	logger.Info("正在关闭 Agent...")
	if err := agent.Stop(); err != nil {
		logger.Error("关闭 Agent 失败: ", err)
		os.Exit(1)
	}

	logger.Info("Agent 已关闭")
}