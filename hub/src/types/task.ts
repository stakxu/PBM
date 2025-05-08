export enum TaskStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled'
}

export enum TaskPriority {
  LOW = 0,
  NORMAL = 1,
  HIGH = 2,
  URGENT = 3
}

export enum TaskType {
  SYSTEM_CHECK = 'system_check',
  PERFORMANCE_TEST = 'performance_test',
  NETWORK_TEST = 'network_test',
  CUSTOM = 'custom'
}

export interface Task {
  id: number;
  agent_uuid: string;
  name: string;
  type: TaskType;
  status: TaskStatus;
  priority: TaskPriority;
  config: any;
  result?: any;
  error?: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
}