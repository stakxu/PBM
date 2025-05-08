import Database from 'better-sqlite3';
import { config } from '../config';
import { Info, Error } from '../logger';
import fs from 'fs';
import path from 'path';
import { initDatabase as initDb } from './init-db';

// 确保数据库目录存在
const dbDir = path.dirname(config.db.path);
if (!fs.existsSync(dbDir)) {
  fs.mkdirSync(dbDir, { recursive: true });
}

// 创建数据库连接
export const db = new Database(config.db.path);

// 初始化数据库
export function initDatabase() {
  try {
    initDb();
    Info('数据库初始化完成');
  } catch (error) {
    Error('数据库初始化失败:', error);
    throw error;
  }
}