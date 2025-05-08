# 数据库设计文档

## 表结构

### agents 表
存储 Agent 的基本信息和状态。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| uuid | TEXT | 主键，Agent 唯一标识 |
| alias | TEXT | Agent 别名 |
| ipv4_address | TEXT | Agent IPv4 地址 |
| ipv6_address | TEXT | Agent IPv6 地址 |
| last_seen | TIMESTAMP | 最后一次心跳时间 |
| status | TEXT | Agent 状态 (online/offline) |
| system_info | TEXT | 静态系统信息(JSON) |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### system_metrics 表
存储系统监控指标数据。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| agent_uuid | TEXT | 关联的 Agent UUID |
| timestamp | TIMESTAMP | 数据采集时间 |
| cpu_usage | REAL | CPU 使用率 |
| memory_used | INTEGER | 已用内存(bytes) |
| memory_total | INTEGER | 总内存(bytes) |
| disk_used | INTEGER | 已用磁盘空间(bytes) |
| disk_total | INTEGER | 总磁盘空间(bytes) |
| network_in | INTEGER | 网络入流量(bytes) |
| network_out | INTEGER | 网络出流量(bytes) |
| tcp_connections | INTEGER | TCP 连接数 |
| udp_connections | INTEGER | UDP 连接数 |
| uptime | REAL | 系统运行时间(秒) |

### tasks 表
存储任务信息。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| agent_uuid | TEXT | 关联的 Agent UUID |
| name | TEXT | 任务名称 |
| type | TEXT | 任务类型 |
| status | TEXT | 任务状态 |
| priority | INTEGER | 任务优先级 |
| config | TEXT | 任务配置(JSON) |
| result | TEXT | 任务结果(JSON) |
| error | TEXT | 错误信息 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |
| started_at | TIMESTAMP | 开始时间 |
| completed_at | TIMESTAMP | 完成时间 |

### task_logs 表
存储任务执行日志。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| task_id | INTEGER | 关联的任务 ID |
| level | TEXT | 日志级别 |
| message | TEXT | 日志内容 |
| created_at | TIMESTAMP | 创建时间 |

## 索引

- agents_status_idx: 用于快速查询特定状态的 Agent
- system_metrics_timestamp_idx: 用于时间范围查询
- system_metrics_agent_idx: 用于查询特定 Agent 的指标
- tasks_status_idx: 用于任务状态查询
- tasks_agent_idx: 用于查询特定 Agent 的任务
- task_logs_task_idx: 用于查询特定任务的日志

## 外键约束

- system_metrics.agent_uuid -> agents.uuid (CASCADE)
- tasks.agent_uuid -> agents.uuid (CASCADE)
- task_logs.task_id -> tasks.id (CASCADE)