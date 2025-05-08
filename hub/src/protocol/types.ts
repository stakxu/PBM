// 消息类型枚举
export enum MessageType {
  AUTH = 'AUTH',           // 认证消息
  HEARTBEAT = 'HEART',     // 心跳消息
  SYSTEM_INFO = 'SINFO',   // 系统信息
  STATIC_INFO = 'STATIC',  // 静态系统信息
  TASK_RESULT = 'TRSLT',  // 任务结果
  TASK_REQUEST = 'TREQ',  // 任务请求
  CONFIG = 'CONFIG',      // 配置更新
}

// 消息头部接口
export interface MessageHeader {
  type: MessageType;
  length: number;
  timestamp: number;
}

// 完整消息接口
export interface Message {
  header: MessageHeader;
  payload: any;
}

// 系统信息接口
export interface SystemInfo {
  networkTraffic: {
    in: number;
    out: number;
  };
  uptime: number;
  cpu: {
    usage: number;
    model: string;
    cores: number;
  };
  memory: {
    total: number;
    used: number;
  };
  disk: {
    total: number;
    used: number;
  };
  swap: {
    total: number;
    used: number;
  };
  network: {
    tcp: number;
    udp: number;
  };
  uuid: string;
  ipv4: string[];
  ipv6: string[];
}