import { Message, MessageHeader, MessageType } from './types';
import { Debug, Error } from '../logger';

export class MessageParser {
  private buffer: Buffer = Buffer.alloc(0);
  private static HEADER_SIZE = 12; // 4(type) + 4(length) + 4(timestamp)

  public append(chunk: Buffer): void {
    this.buffer = Buffer.concat([this.buffer, chunk]);
    Debug(`接收到数据: ${chunk.length} 字节, 当前缓冲区大小: ${this.buffer.length} 字节`);
    Debug(`原始数据: ${chunk.toString('hex')}`);
  }

  public hasCompleteMessage(): boolean {
    if (this.buffer.length < MessageParser.HEADER_SIZE) {
      Debug(`缓冲区大小不足消息头: ${this.buffer.length} < ${MessageParser.HEADER_SIZE}`);
      return false;
    }

    const dataLength = this.buffer.readUInt32BE(4);
    const hasComplete = this.buffer.length >= MessageParser.HEADER_SIZE + dataLength;
    Debug(`消息体长度: ${dataLength}, 当前缓冲区: ${this.buffer.length}, 是否完整: ${hasComplete}`);
    return hasComplete;
  }

  public parseMessage(): Message | null {
    if (!this.hasCompleteMessage()) {
      return null;
    }

    try {
      // 解析消息头
      const typeStr = this.buffer.toString('utf8', 0, 4);
      const length = this.buffer.readUInt32BE(4);
      const timestamp = this.buffer.readUInt32BE(8);

      Debug(`解析消息头 - 类型: ${typeStr}, 长度: ${length}, 时间戳: ${timestamp}`);
      Debug(`消息头原始数据: ${this.buffer.slice(0, MessageParser.HEADER_SIZE).toString('hex')}`);

      const header: MessageHeader = {
        type: typeStr as MessageType,
        length,
        timestamp,
      };

      // 解析消息体
      const payloadBuffer = this.buffer.slice(
        MessageParser.HEADER_SIZE,
        MessageParser.HEADER_SIZE + length
      );
      const payloadStr = payloadBuffer.toString('utf8');
      Debug(`原始消息体: ${payloadStr}`);
      
      let payload;
      try {
        payload = JSON.parse(payloadStr);
        //Debug(`解析后的消息体:`, payload);
      } catch (e) {
        Error(`JSON解析失败: ${e.message}`);
        Debug(`解析失败的消息体内容: ${payloadStr}`);
        return null;
      }

      // 移除已解析的消息
      this.buffer = this.buffer.slice(MessageParser.HEADER_SIZE + length);
      Debug(`移除已解析消息后缓冲区大小: ${this.buffer.length}`);

      return { 
        header, 
        payload 
      };
    } catch (error) {
      Error('解析消息失败:', error);
      Debug('当前缓冲区内容:', this.buffer.toString('hex'));
      // 清空缓冲区以防止错误累积
      this.buffer = Buffer.alloc(0);
      return null;
    }
  }

  public static createMessage(type: MessageType, payload: any): Buffer {
    const payloadStr = JSON.stringify(payload);
    const payloadBuffer = Buffer.from(payloadStr, 'utf8');
    const length = payloadBuffer.length;
    const timestamp = Math.floor(Date.now() / 1000);

    Debug(`创建消息 - 类型: ${type}, 长度: ${length}, 时间戳: ${timestamp}`);
    Debug(`消息体: ${payloadStr}`);

    const buffer = Buffer.alloc(MessageParser.HEADER_SIZE + length);
    
    // 写入消息头
    buffer.write(type, 0, 4);
    buffer.writeUInt32BE(length, 4);
    buffer.writeUInt32BE(timestamp, 8);
    
    // 写入消息体
    payloadBuffer.copy(buffer, MessageParser.HEADER_SIZE);

    Debug(`完整消息内容: ${buffer.toString('hex')}`);
    return buffer;
  }

  public getBufferSize(): number {
    return this.buffer.length;
  }
}