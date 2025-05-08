package protocol

import (
	"encoding/binary"
	"encoding/json"
)

type MessageParser struct {
	buffer []byte
}

func NewMessageParser() *MessageParser {
	return &MessageParser{
		buffer: make([]byte, 0),
	}
}

func (p *MessageParser) Append(data []byte) {
	p.buffer = append(p.buffer, data...)
}

func (p *MessageParser) HasCompleteMessage() bool {
	headerSize := 12 // 4(type) + 4(length) + 4(timestamp)
	if len(p.buffer) < headerSize {
		return false
	}

	length := binary.BigEndian.Uint32(p.buffer[4:8])
	return len(p.buffer) >= headerSize+int(length)
}

func (p *MessageParser) ParseMessage() *Message {
	if !p.HasCompleteMessage() {
		return nil
	}

	headerSize := 12
	msgType := MessageType(string(p.buffer[0:4]))
	length := binary.BigEndian.Uint32(p.buffer[4:8])
	timestamp := binary.BigEndian.Uint32(p.buffer[8:12])

	header := MessageHeader{
		Type:      msgType,
		Length:    length,
		Timestamp: timestamp,
	}

	payloadBytes := p.buffer[headerSize : headerSize+int(length)]
	var payload interface{}

	switch msgType {
	case MessageTypeAuth:
		var auth AuthPayload
		json.Unmarshal(payloadBytes, &auth)
		payload = &auth
	case MessageTypeHeartbeat:
		var heartbeat HeartbeatPayload
		json.Unmarshal(payloadBytes, &heartbeat)
		payload = &heartbeat
	case MessageTypeSystemInfo:
		var sysInfo SystemInfo
		json.Unmarshal(payloadBytes, &sysInfo)
		payload = &sysInfo
	}

	// 移除已解析的消息
	p.buffer = p.buffer[headerSize+int(length):]

	return &Message{
		Header:  header,
		Payload: payload,
	}
}