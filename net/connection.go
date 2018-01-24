package net

import (
	"errors"
	"net"
	"sync"
)

// 用于向上层传递收到的数据包
type dataChan struct {
	data []byte
	conn *Connection
}

// Connection object
type Connection struct {
	remoteAddress string   // 远端地址
	conn          net.Conn // 底层连接

	packetHandler PacketHandler // 包处理器

	DataChan chan []byte

	stopCmdChan chan struct{} // 断开时发送此命令
}

// RemoteAddr function
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) disconnect() {
	c.stopCmdChan <- struct{}{}
}

func (c *Connection) loop(eventChan chan *Event) {
	defer c.conn.Close()

	for {
		select {

		case <-c.stopCmdChan:
			eventChan <- &Event{Type: EventDisconnected, Err: errors.New("stopped"), Conn: c}
			close(c.DataChan)
			return

		default:
			data, err := c.packetHandler.Receive()
			if err != nil {
				eventChan <- &Event{Type: EventDisconnected, Err: err, Conn: c}
				close(c.DataChan)
				return
			}

			if data != nil {
				c.DataChan <- data
			}
		}
	}
}

// Send send data
func (c *Connection) Send(data []byte) error {
	return c.packetHandler.Send(data)
}

// 管理建立的连接
type connections struct {
	connections map[*Connection]*Connection
	mutex       sync.Mutex
	sem         chan struct{}
}

func (conns *connections) init(n uint32) {
	conns.connections = make(map[*Connection]*Connection, n)
	conns.sem = make(chan struct{}, n)
}

func (conns *connections) size() uint32 {
	return uint32(len(conns.connections))
}

func (conns *connections) add(conn *Connection) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()

	conns.connections[conn] = conn
}

func (conns *connections) remove(conn *Connection) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()

	delete(conns.connections, conn)
	conns.release()
}

func (conns *connections) acquire() { conns.sem <- struct{}{} }
func (conns *connections) release() { <-conns.sem }

// CreateTCPClient creates a client object for tcp
func CreateTCPClient() *TCPClient {
	return &TCPClient{}
}

// CreateTCPServer function a server object for tcp
func CreateTCPServer() *TCPServer {
	return &TCPServer{}
}
