package net

import (
	"fmt"
	"net"
	"sync"
)

// Callback interface
type Callback interface {
	Disconnected(conn *Connection, err error)
	Connected(conn *Connection)
	Received(conn *Connection, packet *Packet)
}

// Connection object
type Connection struct {
	remoteAddress string   // 远端地址
	conn          net.Conn // 连接

	stopCmdChan  chan bool
	exitLoopChan chan bool
}

type packetChan struct {
	packet *Packet
	conn   *Connection
}

// RemoteAddr function
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) disconnect() {
	c.stopCmdChan <- true
	<-c.exitLoopChan
}

func (c *Connection) loop(callback Callback, packetsChan chan *packetChan) {
	defer c.conn.Close()

	for {
		select {
		case <-c.stopCmdChan:
			c.exitLoopChan <- true
			return
		default:
			// read buffer

		}
	}
}

type connections struct {
	connections map[*Connection]*Connection
	mutex       sync.Mutex
}

func (conns *connections) init(n uint32) {
	conns.connections = make(map[*Connection]*Connection, n)
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
}

// Test aaa
func Test() {
	fmt.Println("test")
}

// CreateTCPClient creates a client object for tcp
func CreateTCPClient() *TCPClient {
	return &TCPClient{}
}

// CreateTCPServer function
func CreateTCPServer() *TCPServer {
	return &TCPServer{}
}
