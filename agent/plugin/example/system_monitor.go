package main

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"time"
)

type SystemMonitor struct {
	interval    int
	stop        chan struct{}
	initialized bool
}

type Config struct {
	Interval int `json:"interval"`
}

var Plugin SystemMonitor

func (m *SystemMonitor) Init(config json.RawMessage) error {
	if m.initialized {
		return nil
	}

	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}
	
	if cfg.Interval <= 0 {
		cfg.Interval = 60 // 默认60秒
	}
	
	m.interval = cfg.Interval
	m.stop = make(chan struct{})
	m.initialized = true
	return nil
}

func (m *SystemMonitor) Start() error {
	if !m.initialized {
		return fmt.Errorf("插件未初始化")
	}
	
	go m.monitor()
	return nil
}

func (m *SystemMonitor) Stop() error {
	if !m.initialized {
		return nil
	}
	close(m.stop)
	return nil
}

func (m *SystemMonitor) Name() string {
	return "system_monitor"
}

func (m *SystemMonitor) Version() string {
	return "1.0.0"
}

func (m *SystemMonitor) Description() string {
	return "系统资源监控插件"
}

func (m *SystemMonitor) monitor() {
	ticker := time.NewTicker(time.Duration(m.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			m.collectAndReport()
		}
	}
}

func (m *SystemMonitor) collectAndReport() {
	// CPU 使用率
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil {
		fmt.Printf("CPU Usage: %.2f%%\n", cpuPercent[0])
	}

	// 内存使用情况
	if vmStat, err := mem.VirtualMemory(); err == nil {
		fmt.Printf("Memory Usage: %.2f%%\n", vmStat.UsedPercent)
	}

	// 磁盘使用情况
	if parts, err := disk.Partitions(false); err == nil {
		for _, part := range parts {
			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				fmt.Printf("Disk Usage (%s): %.2f%%\n", part.Mountpoint, usage.UsedPercent)
			}
		}
	}

	// 网络连接数
	if conns, err := net.Connections("all"); err == nil {
		tcpCount := 0
		udpCount := 0
		for _, conn := range conns {
			switch conn.Type {
			case "tcp":
				tcpCount++
			case "udp":
				udpCount++
			}
		}
		fmt.Printf("Network Connections - TCP: %d, UDP: %d\n", tcpCount, udpCount)
	}
}