import { db } from '..';
import { logger } from '../../logger';
import fs from 'fs';
import path from 'path';

export async function runMigrations() {
  const migrationsDir = path.join(__dirname);
  const migrationFiles = fs.readdirSync(migrationsDir)
    .filter(file => file.endsWith('.sql'))
    .sort();

  logger.info('开始执行数据库迁移...');

  for (const file of migrationFiles) {
    const migration = fs.readFileSync(path.join(migrationsDir, file), 'utf8');
    try {
      db.exec(migration);
      logger.info(`成功执行迁移: ${file}`);
    } catch (error) {
      logger.error(`执行迁移失败 ${file}:`, error);
      throw error;
    }
  }

  logger.info('数据库迁移完成');
}