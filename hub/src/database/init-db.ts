import { db } from './index';
import { Info } from '../logger';

export function initDatabase() {
  try {
    // 创建 agents 表
    db.exec(`
      CREATE TABLE IF NOT EXISTS agents (
        uuid TEXT PRIMARY KEY,
        alias TEXT,
        ipv4_address TEXT,
        ipv6_address TEXT,
        last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        status TEXT DEFAULT 'offline',
        system_info TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );
      CREATE INDEX IF NOT EXISTS agents_status_idx ON agents(status);
    `);

    // 创建系统指标表
    db.exec(`
      CREATE TABLE IF NOT EXISTS system_metrics (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        agent_uuid TEXT NOT NULL,
        timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        cpu_usage REAL,
        memory_used INTEGER,
        memory_total INTEGER,
        disk_used INTEGER,
        disk_total INTEGER,
        network_in INTEGER,
        network_out INTEGER,
        tcp_connections INTEGER,
        udp_connections INTEGER,
        uptime REAL,
        FOREIGN KEY (agent_uuid) REFERENCES agents(uuid) ON DELETE CASCADE
      );
      CREATE INDEX IF NOT EXISTS system_metrics_timestamp_idx ON system_metrics(timestamp);
      CREATE INDEX IF NOT EXISTS system_metrics_agent_idx ON system_metrics(agent_uuid);
    `);

    // 创建任务表
    db.exec(`
      CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        agent_uuid TEXT NOT NULL,
        name TEXT NOT NULL,
        type TEXT NOT NULL,
        status TEXT DEFAULT 'pending',
        priority INTEGER DEFAULT 0,
        config TEXT,
        result TEXT,
        error TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        started_at TIMESTAMP,
        completed_at TIMESTAMP,
        FOREIGN KEY (agent_uuid) REFERENCES agents(uuid) ON DELETE CASCADE
      );
      CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks(status);
      CREATE INDEX IF NOT EXISTS tasks_agent_idx ON tasks(agent_uuid);
    `);

    // 创建任务日志表
    db.exec(`
      CREATE TABLE IF NOT EXISTS task_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        level TEXT NOT NULL,
        message TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
      );
      CREATE INDEX IF NOT EXISTS task_logs_task_idx ON task_logs(task_id);
    `);

    Info('数据库初始化成功');
  } catch (error) {
    throw error;
  }
}