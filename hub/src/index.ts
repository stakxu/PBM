import { config } from './config';
import { Info, Error } from './logger';
import { initDatabase } from './database';
import { TCPServer } from './tcp/server';
import { AgentManager } from './managers/agent-manager';

async function main() {
  try {
    // 初始化数据库
    initDatabase();
    
    Info('Hub 正在启动...');
    
    // 创建 Agent 管理器
    const agentManager = new AgentManager();
    
    // 启动 TCP 服务器
    const tcpServer = new TCPServer(agentManager);
    tcpServer.start();
    
    // 优雅关闭
    process.on('SIGTERM', () => {
      Info('收到 SIGTERM 信号，正在关闭服务器...');
      tcpServer.stop();
    });

    process.on('SIGINT', () => {
      Info('收到 SIGINT 信号，正在关闭服务器...');
      tcpServer.stop();
      process.exit(0);
    });
    
  } catch (error) {
    Error('Hub 启动失败:', error);
    process.exit(1);
  }
}

// 启动应用
main();