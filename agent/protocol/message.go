package protocol

import (
	"encoding/binary"
	"encoding/json"
	"time"
)

type MessageType string

const (
	MessageTypeAuth       MessageType = "AUTH"
	MessageTypeHeartbeat MessageType = "HEART"
	MessageTypeSystemInfo MessageType = "SINFO"
	MessageTypeStaticInfo MessageType = "STATIC"
	MessageTypeTaskResult MessageType = "TRSLT"
	MessageTypeConfig     MessageType = "CONFIG"
)

type Message struct {
	Header  MessageHeader
	Payload interface{}
}

type MessageHeader struct {
	Type      MessageType
	Length    uint32
	Timestamp uint32
}

type AuthPayload struct {
	Key   string `json:"key"`
	UUID  string `json:"uuid"`
	Alias string `json:"alias"`
}

type HeartbeatPayload struct {
	UUID string `json:"uuid"`
}

// 静态系统信息
type StaticSystemInfo struct {
	UUID     string   `json:"uuid"`
	Alias    string   `json:"alias"`
	CPU      CPUInfo  `json:"cpu"`
	Memory   MemInfo  `json:"memory"`
	Disk     DiskInfo `json:"disk"`
	Swap     SwapInfo `json:"swap"`
	IPv4     []string `json:"ipv4"`
	IPv6     []string `json:"ipv6"`
	UpdateAt int64    `json:"updateAt"`
}

type CPUInfo struct {
	Model string `json:"model"`
	Cores int    `json:"cores"`
}

type MemInfo struct {
	Total uint64 `json:"total"`
}

type DiskInfo struct {
	Total uint64 `json:"total"`
}

type SwapInfo struct {
	Total uint64 `json:"total"`
}

// 动态系统信息
type SystemInfo struct {
	UUID          string `json:"uuid"`
	NetworkTraffic struct {
		In  uint64 `json:"in"`
		Out uint64 `json:"out"`
	} `json:"networkTraffic"`
	Uptime  float64 `json:"uptime"`
	CPU     struct {
		Usage float64 `json:"usage"`
	} `json:"cpu"`
	Memory struct {
		Used uint64 `json:"used"`
	} `json:"memory"`
	Disk struct {
		Used uint64 `json:"used"`
	} `json:"disk"`
	Swap struct {
		Used uint64 `json:"used"`
	} `json:"swap"`
	Network struct {
		TCP int `json:"tcp"`
		UDP int `json:"udp"`
	} `json:"network"`
}

func NewMessage(msgType MessageType, payload interface{}) *Message {
	return &Message{
		Header: MessageHeader{
			Type:      msgType,
			Timestamp: uint32(time.Now().Unix()),
		},
		Payload: payload,
	}
}

func (m *Message) Encode() []byte {
	payloadBytes, _ := json.Marshal(m.Payload)
	m.Header.Length = uint32(len(payloadBytes))

	headerSize := 12 // 4(type) + 4(length) + 4(timestamp)
	data := make([]byte, headerSize+len(payloadBytes))

	// 写入消息类型 (4字节)
	copy(data[0:4], []byte(m.Header.Type))

	// 写入数据长度 (4字节)
	binary.BigEndian.PutUint32(data[4:8], m.Header.Length)

	// 写入时间戳 (4字节)
	binary.BigEndian.PutUint32(data[8:12], m.Header.Timestamp)

	// 写入负载数据
	copy(data[headerSize:], payloadBytes)

	return data
}

func (m *Message) DecodePayload(v interface{}) error {
	payloadBytes, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(payloadBytes, v)
}