package main

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"time"
)

type MemoryMonitor struct {
	interval int
	stop     chan struct{}
}

type Config struct {
	Interval int `json:"interval"`
}

var Plugin MemoryMonitor

func (m *MemoryMonitor) Init(config json.RawMessage) error {
	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}
	
	if cfg.Interval <= 0 {
		cfg.Interval = 60 // 默认60秒
	}
	
	m.interval = cfg.Interval
	m.stop = make(chan struct{})
	return nil
}

func (m *MemoryMonitor) Start() error {
	go m.monitor()
	return nil
}

func (m *MemoryMonitor) Stop() error {
	close(m.stop)
	return nil
}

func (m *MemoryMonitor) Name() string {
	return "memory_monitor"
}

func (m *MemoryMonitor) monitor() {
	ticker := time.NewTicker(time.Duration(m.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			if v, err := mem.VirtualMemory(); err == nil {
				fmt.Printf("Memory Usage: %.2f%%\n", v.UsedPercent)
			}
		}
	}
}