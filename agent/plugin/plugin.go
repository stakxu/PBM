package plugin

import (
	"agent/logger"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// Plugin 接口定义了插件必须实现的方法
type Plugin interface {
	Init(config json.RawMessage) error
	Start() error
	Stop() error
	Name() string
	Version() string
	Description() string
}

// PluginInfo 存储插件元数据
type PluginInfo struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Config      json.RawMessage `json:"config"`
}

// Manager 管理插件的生命周期
type Manager struct {
	plugins     map[string]Plugin
	configs     map[string]json.RawMessage
	pluginDir   string
	mutex       sync.RWMutex
	initialized bool
}

func NewManager() *Manager {
	return &Manager{
		plugins:   make(map[string]Plugin),
		configs:   make(map[string]json.RawMessage),
		pluginDir: "plugins",
	}
}

// Init 初始化插件管理器
func (m *Manager) Init() error {
	if m.initialized {
		return nil
	}

	// 创建插件目录
	if err := os.MkdirAll(m.pluginDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %v", err)
	}

	// 加载插件配置
	if err := m.loadConfigs(); err != nil {
		return err
	}

	// 加载所有插件
	if err := m.loadAllPlugins(); err != nil {
		return err
	}

	m.initialized = true
	return nil
}

// loadConfigs 加载插件配置文件
func (m *Manager) loadConfigs() error {
	configPath := filepath.Join(m.pluginDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取插件配置失败: %v", err)
	}

	var configs map[string]json.RawMessage
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("解析插件配置失败: %v", err)
	}

	m.configs = configs
	return nil
}

// loadAllPlugins 加载所有插件
func (m *Manager) loadAllPlugins() error {
	files, err := os.ReadDir(m.pluginDir)
	if err != nil {
		return fmt.Errorf("读取插件目录失败: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".so" {
			continue
		}

		if err := m.Load(filepath.Join(m.pluginDir, file.Name())); err != nil {
			logger.Error("加载插件失败:", file.Name(), err)
		}
	}

	return nil
}

// Load 加载单个插件
func (m *Manager) Load(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 加载插件
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("加载插件失败: %v", err)
	}

	// 获取插件符号
	symPlugin, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("查找插件符号失败: %v", err)
	}

	// 类型断言
	plugin, ok := symPlugin.(Plugin)
	if !ok {
		return fmt.Errorf("插件类型断言失败")
	}

	// 获取插件名称
	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("插件 %s 已存在", name)
	}

	// 初始化插件
	if config, exists := m.configs[name]; exists {
		if err := plugin.Init(config); err != nil {
			return fmt.Errorf("初始化插件失败: %v", err)
		}
	}

	m.plugins[name] = plugin
	logger.Info("插件加载成功:", name, "版本:", plugin.Version())
	return nil
}

// Start 启动指定插件
func (m *Manager) Start(name string) error {
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}

	if err := plugin.Start(); err != nil {
		return fmt.Errorf("启动插件失败: %v", err)
	}

	logger.Info("插件启动成功:", name)
	return nil
}

// Stop 停止指定插件
func (m *Manager) Stop(name string) error {
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("插件 %s 不存在", name)
	}

	if err := plugin.Stop(); err != nil {
		return fmt.Errorf("停止插件失败: %v", err)
	}

	logger.Info("插件停止成功:", name)
	return nil
}

// StopAll 停止所有插件
func (m *Manager) StopAll() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, plugin := range m.plugins {
		if err := plugin.Stop(); err != nil {
			logger.Error("停止插件失败:", name, err)
		}
	}
}

// GetPlugin 获取指定插件
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	plugin, exists := m.plugins[name]
	return plugin, exists
}

// ListPlugins 列出所有已加载的插件
func (m *Manager) ListPlugins() []PluginInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugins := make([]PluginInfo, 0, len(m.plugins))
	for name, p := range m.plugins {
		plugins = append(plugins, PluginInfo{
			Name:        name,
			Version:     p.Version(),
			Description: p.Description(),
			Config:      m.configs[name],
		})
	}
	return plugins
}