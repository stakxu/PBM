import { EventEmitter } from 'events';
import { db } from '../database';
import { logger } from '../logger';
import { Task, TaskStatus, TaskPriority, TaskType } from '../types/task';
import { MessageParser } from '../protocol/parser';
import { MessageType } from '../protocol/types';

export class TaskManager extends EventEmitter {
  constructor() {
    super();
  }

  public async createTask(task: Omit<Task, 'id' | 'status' | 'created_at' | 'updated_at'>): Promise<number> {
    const stmt = db.prepare(`
      INSERT INTO tasks (
        agent_uuid, name, type, priority, config,
        status, created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `);

    const result = stmt.run(
      task.agent_uuid,
      task.name,
      task.type,
      task.priority || TaskPriority.NORMAL,
      JSON.stringify(task.config),
      TaskStatus.PENDING
    );

    logger.info(`创建任务成功: ${result.lastInsertRowid}`);
    return result.lastInsertRowid as number;
  }

  public async getTask(id: number): Promise<Task | null> {
    const stmt = db.prepare('SELECT * FROM tasks WHERE id = ?');
    return stmt.get(id) as Task | null;
  }

  public async updateTaskStatus(id: number, status: TaskStatus, result?: any, error?: string): Promise<void> {
    const stmt = db.prepare(`
      UPDATE tasks
      SET status = ?,
          result = ?,
          error = ?,
          updated_at = CURRENT_TIMESTAMP,
          ${status === TaskStatus.RUNNING ? 'started_at = CURRENT_TIMESTAMP,' : ''}
          ${status === TaskStatus.COMPLETED || status === TaskStatus.FAILED ? 'completed_at = CURRENT_TIMESTAMP,' : ''}
          updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `);

    stmt.run(status, result ? JSON.stringify(result) : null, error, id);
    logger.info(`更新任务状态: ${id} -> ${status}`);
  }

  public async getPendingTasks(agentUuid: string): Promise<Task[]> {
    const stmt = db.prepare(`
      SELECT * FROM tasks 
      WHERE agent_uuid = ? AND status = ? 
      ORDER BY priority DESC, created_at ASC
    `);
    return stmt.all(agentUuid, TaskStatus.PENDING) as Task[];
  }

  public async addTaskLog(taskId: number, level: string, message: string): Promise<void> {
    const stmt = db.prepare(`
      INSERT INTO task_logs (task_id, level, message, created_at)
      VALUES (?, ?, ?, CURRENT_TIMESTAMP)
    `);
    stmt.run(taskId, level, message);
  }

  public createTaskMessage(task: Task): Buffer {
    return MessageParser.createMessage(MessageType.TASK_REQUEST, {
      taskId: task.id,
      type: task.type,
      config: task.config
    });
  }
}