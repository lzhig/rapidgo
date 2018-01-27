package net

import (
	"bufio"
	"net"
)

// // ICallback interface
// type ICallback interface {
// 	Disconnected(conn *Connection, err error)
// 	Connected(conn *Connection)
// 	Received(conn *Connection, data []byte)
// }

// Config 用于初始化网络引擎
type Config struct {
	PacketHandlerFactory func(net.Conn) PacketHandler
}

var config = &Config{
	PacketHandlerFactory: func(c net.Conn) PacketHandler {
		return &defaultPacketHandler{
			conn:      c,
			bufReader: bufio.NewReader(c),
			bufWriter: bufio.NewWriter(c),
		}
	},
}

// Init 初始化
func Init(cfg *Config) {
	config = cfg
}

// EventType 通知上层事件类型
type EventType int

const (
	// EventConnected 连接成功
	EventConnected = iota

	// EventDisconnected 断开
	EventDisconnected

	// EventSendFailed 发送错误
	EventSendFailed
)

// Event 事件
type Event struct {
	Type EventType
	Err  error
	Conn *Connection
}
