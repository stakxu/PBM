import { EventEmitter } from 'events';
import { db } from '../database';
import { Debug, Info, Error } from '../logger';
import { AgentStatus, AgentInfo, AgentConfig, AgentEventType, AgentEvent } from '../types/agent';
import { MessageParser } from '../protocol/parser';
import { MessageType } from '../protocol/types';

export class AgentManager extends EventEmitter {
  private agents: Map<string, AgentInfo> = new Map();
  private defaultConfig: AgentConfig = {
    systemInfoInterval: 60,
    heartbeatInterval: 30,
    reconnectInterval: 5,
  };

  constructor() {
    super();
    this.loadAgentsFromDatabase();
  }

  private async loadAgentsFromDatabase(): Promise<void> {
    try {
      const stmt = db.prepare('SELECT * FROM agents');
      const agents = stmt.all();
      
      for (const agent of agents) {
        this.agents.set(agent.uuid, {
          uuid: agent.uuid,
          ipv4Address: agent.ipv4_address,
          ipv6Address: agent.ipv6_address,
          status: agent.status as AgentStatus,
          lastSeen: new Date(agent.last_seen),
          config: this.defaultConfig,
          systemInfo: agent.system_info ? JSON.parse(agent.system_info) : undefined,
        });
      }
      
      Info(`已从数据库加载 ${agents.length} 个 Agent 信息`);
    } catch (error) {
      Error('从数据库加载 Agent 信息失败:', error);
    }
  }

  public getAgent(uuid: string): AgentInfo | undefined {
    return this.agents.get(uuid);
  }

  public getAllAgents(): AgentInfo[] {
    return Array.from(this.agents.values());
  }

  public getOnlineAgents(): AgentInfo[] {
    return this.getAllAgents().filter(agent => agent.status === AgentStatus.ONLINE);
  }

  public updateAgentStatus(uuid: string, status: AgentStatus): void {
    const agent = this.agents.get(uuid);
    if (agent) {
      agent.status = status;
      agent.lastSeen = new Date();
      
      // 更新数据库
      const stmt = db.prepare(`
        UPDATE agents 
        SET status = ?, last_seen = CURRENT_TIMESTAMP 
        WHERE uuid = ?
      `);
      stmt.run(status, uuid);

      // 触发状态变更事件
      this.emit(AgentEventType.STATUS_CHANGED, {
        type: AgentEventType.STATUS_CHANGED,
        agentUuid: uuid,
        data: { status, lastSeen: agent.lastSeen }
      } as AgentEvent);

      Info(`Agent ${uuid} 状态更新为 ${status}`);
    }
  }

  public updateAgentSystemInfo(uuid: string, systemInfo: any): void {
    const agent = this.agents.get(uuid);
    if (agent) {
      agent.systemInfo = systemInfo;
      agent.lastSeen = new Date();

      // 更新数据库
      const stmt = db.prepare(`
        UPDATE agents 
        SET system_info = ?, last_seen = CURRENT_TIMESTAMP 
        WHERE uuid = ?
      `);
      stmt.run(JSON.stringify(systemInfo), uuid);

      // 触发系统信息更新事件
      this.emit(AgentEventType.SYSTEM_INFO_UPDATED, {
        type: AgentEventType.SYSTEM_INFO_UPDATED,
        agentUuid: uuid,
        data: systemInfo
      } as AgentEvent);

      Info(`已更新 Agent ${uuid} 的系统信息`);
    }
  }

  public updateAgentConfig(uuid: string, config: Partial<AgentConfig>): void {
    const agent = this.agents.get(uuid);
    if (agent) {
      agent.config = { ...agent.config, ...config };

      // 触发配置更新事件
      this.emit(AgentEventType.CONFIG_UPDATED, {
        type: AgentEventType.CONFIG_UPDATED,
        agentUuid: uuid,
        data: agent.config
      } as AgentEvent);

      Info(`已更新 Agent ${uuid} 的配置`);
    }
  }

  public registerAgent(uuid: string, ipv4Address?: string, ipv6Address?: string): void {
    if (!this.agents.has(uuid)) {
      const agentInfo: AgentInfo = {
        uuid,
        ipv4Address,
        ipv6Address,
        status: AgentStatus.ONLINE,
        lastSeen: new Date(),
        config: this.defaultConfig,
      };

      this.agents.set(uuid, agentInfo);

      // 更新数据库
      const stmt = db.prepare(`
        INSERT INTO agents (uuid, ipv4_address, ipv6_address, status, last_seen)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(uuid) DO UPDATE SET
        ipv4_address = excluded.ipv4_address,
        ipv6_address = excluded.ipv6_address,
        status = excluded.status,
        last_seen = excluded.last_seen
      `);
      stmt.run(uuid, ipv4Address, ipv6Address, AgentStatus.ONLINE);

      Info(`新 Agent 注册成功: ${uuid} (IPv4: ${ipv4Address}, IPv6: ${ipv6Address})`);
    }
  }

  public removeAgent(uuid: string): void {
    if (this.agents.has(uuid)) {
      this.agents.delete(uuid);
      
      // 从数据库中删除
      const stmt = db.prepare('DELETE FROM agents WHERE uuid = ?');
      stmt.run(uuid);

      Info(`Agent ${uuid} 已被移除`);
    }
  }

  public sendConfigToAgent(uuid: string): Buffer {
    const agent = this.agents.get(uuid);
    if (!agent) {
      throw new Error(`Agent ${uuid} 不存在`);
    }

    return MessageParser.createMessage(MessageType.CONFIG, agent.config);
  }
}