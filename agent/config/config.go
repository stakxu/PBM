package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Hub struct {
		Address         string   `yaml:"address"`
		BackupAddresses []string `yaml:"backup_addresses"`
		Port           int      `yaml:"port"`
		Protocol       string   `yaml:"protocol"` // ipv4, ipv6, auto
	} `yaml:"hub"`
	Auth struct {
		Key string `yaml:"key"`
	} `yaml:"auth"`
	Agent struct {
		Alias              string `yaml:"alias"`              // Agent 别名
		SystemInfoInterval int    `yaml:"systemInfoInterval"` // 系统信息上报间隔（秒）
		StaticInfoInterval int    `yaml:"staticInfoInterval"` // 静态信息重新上报时间（小时）
		HeartbeatInterval int    `yaml:"heartbeatInterval"`  // 心跳间隔（秒）
		ReconnectInterval int    `yaml:"reconnectInterval"`  // 重连间隔（秒）
	} `yaml:"agent"`
	Log struct {
		Level string `yaml:"level"`
		Path  string `yaml:"path"`
	} `yaml:"log"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}