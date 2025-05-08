import winston from 'winston';
import { config } from '../config';
import path from 'path';
import fs from 'fs';

// 确保日志目录存在
const logDir = path.dirname(config.log.path);
if (!fs.existsSync(logDir)) {
  fs.mkdirSync(logDir, { recursive: true });
}

// 控制台日志格式
const consoleFormat = winston.format.printf(({ level, message, timestamp }) => {
  return `[${timestamp}] [${level.toUpperCase().padEnd(5)}] ${message}`;
});

// 文件日志格式
const fileFormat = winston.format.printf(({ level, message, timestamp }) => {
  return `[${timestamp}] [${level.toUpperCase().padEnd(5)}] ${message}`;
});

// 创建日志记录器
const logger = winston.createLogger({
  level: config.log.level,
  format: winston.format.combine(
    winston.format.timestamp({
      format: config.log.timeFormat
    }),
    fileFormat
  ),
  transports: [
    // 控制台输出
    new winston.transports.Console({
      level: config.log.level,
      format: winston.format.combine(
        winston.format.timestamp({
          format: config.log.timeFormat
        }),
        winston.format.colorize(),
        consoleFormat
      ),
    }),
    // 文件输出
    new winston.transports.File({
      filename: config.log.path,
      level: config.log.level,
    })
  ],
});

// 导出标准化的日志方法
export function Debug(...args: any[]) {
  const message = formatMessage(args);
  logger.debug(message);
}

export function Info(...args: any[]) {
  const message = formatMessage(args);
  logger.info(message);
}

export function Warn(...args: any[]) {
  const message = formatMessage(args);
  logger.warn(message);
}

export function Error(...args: any[]) {
  const message = formatMessage(args);
  logger.error(message);
}

// 格式化消息
function formatMessage(args: any[]): string {
  return args.map(arg => {
    if (typeof arg === 'object') {
      return JSON.stringify(arg, null, 2);
    }
    return String(arg);
  }).join(' ');
}

// 导出 logger 实例供直接使用
export { logger };