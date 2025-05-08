// Agent 状态枚举
export enum AgentStatus {
  ONLINE = 'online',
  OFFLINE = 'offline',
  DISCONNECTED = 'disconnected',
}

// Agent 配置接口
export interface AgentConfig {
  systemInfoInterval: number;  // 系统信息上报间隔（秒）
  heartbeatInterval: number;   // 心跳间隔（秒）
  reconnectInterval: number;   // 重连间隔（秒）
}

// Agent 信息接口
export interface AgentInfo {
  uuid: string;
  ipv4Address?: string;
  ipv6Address?: string;
  status: AgentStatus;
  lastSeen: Date;
  config: AgentConfig;
  systemInfo?: any;
}

// Agent 事件类型
export enum AgentEventType {
  STATUS_CHANGED = 'status_changed',
  SYSTEM_INFO_UPDATED = 'system_info_updated',
  CONFIG_UPDATED = 'config_updated',
}

// Agent 事件接口
export interface AgentEvent {
  type: AgentEventType;
  agentUuid: string;
  data: any;
}