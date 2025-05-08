# 系统信息采集与监控平台

这是一个基于 Go 和 Node.js 开发的分布式系统信息采集与监控平台,由 Hub 和 Agent 两部分组成。Agent 采用插件化设计,支持动态加载功能模块,通过 TCP 长连接与 Hub 进行双向通信。Hub 作为平台的中心节点,负责接收、存储和管理来自所有 Agent 的数据。

## 功能特点

- 实时系统信息采集
  - CPU 使用率
  - 内存使用情况 
  - 磁盘使用情况
  - 网络流量统计
  - 系统运行时间
  - 网络连接数等

- 插件化任务系统
  - 动态加载/卸载插件
  - 自定义监控任务
  - 性能测试任务
  - 插件生命周期管理

- 安全可靠
  - 密钥认证
  - 心跳检测
  - 自动重连
  - 数据持久化

## 系统要求

### Agent
- 操作系统: Windows, Linux (x86_64)
- 运行环境: Go 1.21+
- 最低配置: 1核CPU, 256MB内存, 1GB硬盘

### Hub  
- 操作系统: Windows, Linux (x86_64)
- 运行环境: Node.js 18+
- 推荐配置: 2核CPU, 1GB内存, 5GB硬盘

## 快速开始

### 安装 Hub

1. 克隆代码并安装依赖:

```bash
git clone <repository_url>
cd hub
npm install
```

2. 配置环境变量:

```bash
cp .env.example .env
```

编辑 .env 文件:

```
HUB_PORT=3000
TCP_PORT=3001
AUTH_KEY=your-secret-key
LOG_LEVEL=info
DB_PATH=data/hub.db
```

3. 启动 Hub:

```bash
npm run dev
```

### 安装 Agent 

1. 克隆代码:

```bash
git clone <repository_url>
cd agent
```

2. 修改配置文件 config.yaml:

```yaml
hub:
  address: "localhost"  # Hub 服务器地址
  port: 3001           # TCP 服务端口
  protocol: "auto"     # 连接协议(ipv4/ipv6/auto)

auth:
  key: "your-secret-key"  # 与 Hub 相同的认证密钥

agent:
  alias: ""              # Agent 别名
  systemInfoInterval: 60 # 系统信息上报间隔(秒)
  heartbeatInterval: 30  # 心跳间隔(秒)
  reconnectInterval: 5   # 重连间隔(秒)

log:
  level: "info"         # 日志级别
  path: "logs/agent.log" # 日志文件路径
```

3. 编译并运行:

```bash
go build
./agent
```

## 插件开发

Agent 支持通过插件扩展功能。插件需要实现以下接口:

```go
type Plugin interface {
    Init(config json.RawMessage) error
    Start() error
    Stop() error
    Name() string
    Version() string
    Description() string
}
```

插件示例可参考 `agent/plugin/example/` 目录。

## 项目结构

```
.
├── agent/                 # Agent 源码
│   ├── config/           # 配置管理
│   ├── core/             # 核心功能
│   ├── logger/           # 日志模块
│   ├── plugin/           # 插件系统
│   └── protocol/         # 通信协议
│
└── hub/                  # Hub 源码
    ├── src/
    │   ├── config/      # 配置管理
    │   ├── database/    # 数据库操作
    │   ├── logger/      # 日志模块
    │   ├── managers/    # 业务管理
    │   ├── protocol/    # 通信协议
    │   └── tcp/         # TCP 服务
    └── docs/            # 文档
```

## API 文档

### TCP 消息类型

- AUTH: 认证消息
- HEARTBEAT: 心跳消息
- SYSTEM_INFO: 系统信息
- TASK_RESULT: 任务结果
- TASK_REQUEST: 任务请求

详细协议文档请参考 `docs/protocol.md`。

## 数据库设计

系统使用 SQLite 数据库存储数据,主要包含以下表:

- agents: Agent 基本信息
- system_metrics: 系统指标数据
- tasks: 任务信息
- task_logs: 任务执行日志

详细表结构请参考 `docs/database.md`。

## 开发计划

详见 TODO.md 文件。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 LICENSE 文件