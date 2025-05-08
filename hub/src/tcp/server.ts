import net from 'net';
import { config } from '../config';
import { Debug, Info, Warn, Error } from '../logger';
import { MessageParser } from '../protocol/parser';
import { Message, MessageType } from '../protocol/types';
import { AgentManager } from '../managers/agent-manager';
import { db } from '../database';

export class TCPServer {
  private server: net.Server;
  private server6?: net.Server;
  private clients: Map<string, net.Socket> = new Map();
  private parsers: Map<string, MessageParser> = new Map();
  private agentManager: AgentManager;
  private authenticatedClients: Set<string> = new Set();

  constructor(agentManager: AgentManager) {
    this.server = net.createServer(this.handleConnection.bind(this));
    if (config.hub.enableIPv6) {
      this.server6 = net.createServer(this.handleConnection.bind(this));
    }
    this.agentManager = agentManager;

    // 添加服务器事件监听
    this.server.on('error', (error) => {
      Error('TCP IPv4 服务器错误:', error);
    });

    this.server.on('listening', () => {
      const address = this.server.address() as net.AddressInfo;
      Info(`TCP IPv4 服务器开始监听 ${address.address}:${address.port}`);
    });

    if (this.server6) {
      this.server6.on('error', (error) => {
        Error('TCP IPv6 服务器错误:', error);
      });

      this.server6.on('listening', () => {
        const address = this.server6?.address() as net.AddressInfo;
        Info(`TCP IPv6 服务器开始监听 [${address.address}]:${address.port}`);
      });
    }
  }

  public start(): void {
    Info('正在启动 TCP 服务器...');
    
    try {
      // 启动 IPv4 服务器
      this.server.listen(config.hub.tcpPort, config.hub.address, () => {
        const address = this.server.address() as net.AddressInfo;
        Info(`TCP IPv4 服务器正在监听 ${address.address}:${address.port}`);
      });

      // 如果启用了 IPv6，启动 IPv6 服务器
      if (this.server6 && config.hub.enableIPv6) {
        // 设置 IPv6 only 以避免地址冲突
        this.server6.listen(config.hub.tcpPort, config.hub.address6, () => {
          const address = this.server6?.address() as net.AddressInfo;
          Info(`TCP IPv6 服务器正在监听 [${address.address}]:${address.port}`);
        });
      }
    } catch (error) {
      Error('启动 TCP 服务器失败:', error);
      throw error;
    }
  }

  private handleConnection(socket: net.Socket): void {
    const clientId = `${socket.remoteAddress}:${socket.remotePort}`;
    Info(`新的客户端连接: ${clientId}`);
    Debug(`客户端连接详情 - 本地地址: ${socket.localAddress}:${socket.localPort}, 远程地址: ${socket.remoteAddress}:${socket.remotePort}`);

    // 设置 TCP keepalive
    socket.setKeepAlive(true, 30000);
    Debug(`已为客户端 ${clientId} 设置 keepalive`);

    // 初始化时给予充足的认证时间
    socket.setTimeout(120000);
    Debug(`已为客户端 ${clientId} 设置初始超时时间: 120秒`);

    this.clients.set(clientId, socket);
    this.parsers.set(clientId, new MessageParser());
    Debug(`已创建客户端 ${clientId} 的解析器和连接记录`);

    socket.on('data', (data) => {
      Info(`收到来自 ${clientId} 的数据，长度: ${data.length} 字节`);
      Debug(`原始数据内容: ${data.toString('hex')}`);
      try {
        this.handleData(clientId, data);
      } catch (error) {
        Error(`处理客户端 ${clientId} 数据时发生错误:`, error);
      }
    });
    
    socket.on('close', (hadError) => {
      Info(`客户端 ${clientId} 连接关闭${hadError ? ' (发生错误)' : ''}`);
      this.handleDisconnect(clientId);
    });

    socket.on('error', (error) => {
      Error(`客户端 ${clientId} 连接错误:`, error);
      this.handleError(clientId, error);
    });

    socket.on('timeout', () => {
      if (this.authenticatedClients.has(clientId)) {
        Warn(`已认证的客户端 ${clientId} 连接超时`);
        socket.end();
      } else {
        Warn(`未认证的客户端 ${clientId} 认证超时`);
        socket.destroy();
      }
    });
  }

  private handleData(clientId: string, data: Buffer): void {
    const parser = this.parsers.get(clientId);
    if (!parser) {
      Warn(`未找到客户端 ${clientId} 的解析器`);
      return;
    }

    Debug(`开始处理来自 ${clientId} 的数据:
    - 数据长度: ${data.length} 字节
    - 原始数据: ${data.toString('hex')}
    - UTF8 内容: ${data.toString('utf8')}`);

    try {
      parser.append(data);
      Debug(`数据已添加到解析器缓冲区，当前缓冲区大小: ${parser.getBufferSize()} 字节`);

      while (parser.hasCompleteMessage()) {
        const message = parser.parseMessage();
        if (message) {
          Debug(`解析到完整消息:
          - 类型: ${message.header.type}
          - 长度: ${message.header.length}
          - 时间戳: ${message.header.timestamp}
          - 负载: ${JSON.stringify(message.payload, null, 2)}`);
          
          this.handleMessage(clientId, message);
        } else {
          Warn(`解析消息失败: ${clientId}`);
        }
      }
    } catch (error) {
      Error(`处理客户端 ${clientId} 数据时发生错误:`, error);
      // 清理连接
      const socket = this.clients.get(clientId);
      if (socket) {
        socket.destroy();
      }
    }
  }

  private handleMessage(clientId: string, message: Message): void {
    Info(`收到来自 ${clientId} 的消息类型: ${message.header.type}`);
    Debug(`消息详情:
    - 类型: ${message.header.type}
    - 时间戳: ${new Date(message.header.timestamp * 1000).toISOString()}
    - 内容: ${JSON.stringify(message.payload, null, 2)}`);

    try {
      switch (message.header.type) {
        case MessageType.AUTH:
          this.handleAuth(clientId, message);
          break;
        case MessageType.HEARTBEAT:
          this.handleHeartbeat(clientId, message);
          break;
        case MessageType.SYSTEM_INFO:
          this.handleSystemInfo(clientId, message);
          break;
        case MessageType.TASK_RESULT:
          this.handleTaskResult(clientId, message);
          break;
        default:
          Warn(`未知的消息类型: ${message.header.type}`);
      }
    } catch (error) {
      Error(`处理消息时发生错误 - 客户端: ${clientId}, 类型: ${message.header.type}:`, error);
    }
  }

  private handleAuth(clientId: string, message: Message): void {
    Debug(`处理认证消息 - 客户端: ${clientId}
    - 负载: ${JSON.stringify(message.payload, null, 2)}`);
    
    const { key, uuid, alias } = message.payload;
    
    if (key === config.auth.key) {
      Info(`客户端 ${clientId} (UUID: ${uuid}, Alias: ${alias}) 认证成功`);
      
      const socket = this.clients.get(clientId);
      if (socket) {
        // 将客户端标记为已认证
        this.authenticatedClients.add(clientId);
        Debug(`客户端 ${clientId} 已添加到认证列表`);
        
        // 认证成功后设置正常的超时时间
        socket.setTimeout(60000);
        Debug(`已更新客户端 ${clientId} 的超时时间为 60 秒`);
        
        const ipAddress = clientId.split(':')[0];
        this.agentManager.registerAgent(uuid, ipAddress);
        Debug(`已注册 Agent: ${uuid} (IP: ${ipAddress})`);
        
        // 发送配置给 Agent
        try {
          const configMessage = this.agentManager.sendConfigToAgent(uuid);
          Debug(`准备发送配置消息到 Agent ${uuid}:
          - 消息长度: ${configMessage.length} 字节
          - 消息内容: ${configMessage.toString('hex')}`);
          
          socket.write(configMessage, (error) => {
            if (error) {
              Error(`发送配置消息到 Agent ${uuid} 失败:`, error);
            } else {
              Debug(`配置消息已成功发送到 Agent ${uuid}`);
            }
          });
        } catch (error) {
          Error(`准备配置消息时发生错误 - Agent ${uuid}:`, error);
        }
      }
    } else {
      Warn(`客户端 ${clientId} 认证失败: 密钥不匹配
      - 期望的密钥: ${config.auth.key}
      - 收到的密钥: ${key}`);
      const socket = this.clients.get(clientId);
      if (socket) {
        socket.destroy();
      }
    }
  }

  private handleHeartbeat(clientId: string, message: Message): void {
    if (!this.authenticatedClients.has(clientId)) {
      Warn(`未认证的客户端 ${clientId} 发送心跳`);
      return;
    }

    const { uuid } = message.payload;
    Debug(`收到来自 ${uuid} 的心跳消息 - 客户端: ${clientId}`);
    this.agentManager.updateAgentStatus(uuid, 'online');
  }

  private handleSystemInfo(clientId: string, message: Message): void {
    if (!this.authenticatedClients.has(clientId)) {
      Warn(`未认证的客户端 ${clientId} 发送系统信息`);
      return;
    }

    const systemInfo = message.payload;
    this.agentManager.updateAgentSystemInfo(systemInfo.uuid, systemInfo);
    
    Debug(`收到系统信息 [${systemInfo.uuid}]:
      CPU: ${systemInfo.cpu.usage.toFixed(2)}% (${systemInfo.cpu.cores}核)
      内存: ${(systemInfo.memory.used / 1024 / 1024).toFixed(2)}MB / ${(systemInfo.memory.total / 1024 / 1024).toFixed(2)}MB
      磁盘: ${(systemInfo.disk.used / 1024 / 1024 / 1024).toFixed(2)}GB / ${(systemInfo.disk.total / 1024 / 1024 / 1024).toFixed(2)}GB
      网络连接: TCP ${systemInfo.network.tcp}, UDP ${systemInfo.network.udp}
      流量: ↑${(systemInfo.networkTraffic.out / 1024).toFixed(2)}KB ↓${(systemInfo.networkTraffic.in / 1024).toFixed(2)}KB
      运行时间: ${(systemInfo.uptime / 3600).toFixed(2)}小时`);
  }

  private handleTaskResult(clientId: string, message: Message): void {
    if (!this.authenticatedClients.has(clientId)) {
      Warn(`未认证的客户端 ${clientId} 发送任务结果`);
      return;
    }

    const { taskId, result, uuid } = message.payload;
    Debug(`收到任务结果 - TaskID: ${taskId}, UUID: ${uuid}
    结果: ${JSON.stringify(result, null, 2)}`);
    
    try {
      // 更新任务结果
      const stmt = db.prepare(`
        UPDATE tasks
        SET status = 'completed',
            result = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = ? AND agent_uuid = ?
      `);
      
      stmt.run(JSON.stringify(result), taskId, uuid);
      Debug(`已更新任务 ${taskId} 的结果到数据库`);
    } catch (error) {
      Error(`更新任务 ${taskId} 结果失败:`, error);
    }
  }

  private handleDisconnect(clientId: string): void {
    Info(`客户端断开连接: ${clientId}`);
    
    try {
      // 查找对应的 Agent 并更新状态
      const agents = this.agentManager.getAllAgents();
      const agent = agents.find(a => a.ipv4Address === clientId.split(':')[0]);
      if (agent) {
        this.agentManager.updateAgentStatus(agent.uuid, 'offline');
        Debug(`已更新 Agent ${agent.uuid} 状态为 offline`);
      }

      this.authenticatedClients.delete(clientId);
      this.clients.delete(clientId);
      this.parsers.delete(clientId);
      Debug(`已清理客户端 ${clientId} 的所有相关资源`);
    } catch (error) {
      Error(`处理客户端 ${clientId} 断开连接时发生错误:`, error);
    }
  }

  private handleError(clientId: string, error: Error): void {
    Error(`客户端 ${clientId} 发生错误:`, error);
    try {
      const socket = this.clients.get(clientId);
      if (socket) {
        socket.destroy();
        Debug(`已销毁客户端 ${clientId} 的 socket 连接`);
      }
    } catch (error) {
      Error(`处理客户端 ${clientId} 错误时发生异常:`, error);
    }
  }

  public stop(): void {
    Info('正在停止 TCP 服务器...');
    
    try {
      this.server.close(() => {
        Info('TCP IPv4 服务器已关闭');
      });

      if (this.server6) {
        this.server6.close(() => {
          Info('TCP IPv6 服务器已关闭');
        });
      }

      // 关闭所有客户端连接
      for (const [clientId, socket] of this.clients) {
        Debug(`正在关闭客户端 ${clientId} 的连接`);
        socket.destroy();
        this.clients.delete(clientId);
        this.parsers.delete(clientId);
        this.authenticatedClients.delete(clientId);
      }

      Info('所有客户端连接已关闭');
    } catch (error) {
      Error('停止 TCP 服务器时发生错误:', error);
    }
  }
}