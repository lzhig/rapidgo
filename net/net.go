package net

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ICallback interface
type ICallback interface {
	Disconnected(conn *Connection, err error)
	Connected(conn *Connection)
	Received(conn *Connection, packet Packet)
}

// 用于向上层传递收到的数据包
type packetChan struct {
	packet Packet
	conn   *Connection
}

// Connection object
type Connection struct {
	remoteAddress string   // 远端地址
	conn          net.Conn // 底层连接

	stopCmdChan  chan bool // 断开时发送此命令
	exitLoopChan chan bool // 当Connection退出loop时，会传入值
}

// RemoteAddr function
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) disconnect() {
	c.stopCmdChan <- true
	<-c.exitLoopChan
}

func (c *Connection) loop(callback ICallback, packetsChan chan *packetChan) {
	defer c.conn.Close()

	var packet Packet

	i := 0
	s := time.Now()

	defer func() {
		d := time.Now().Sub(s)
		sec := d.Seconds()

		fmt.Println(float64(i) / sec * 0.0625)

	}()

	for {
		select {
		case <-c.stopCmdChan:
			c.exitLoopChan <- true
			return
		default:
			if packet == nil {
				packet = packetFactory.CreatePacket()
			}

			ok, err := packet.FillFrom(c.conn)
			if err != nil && err != io.EOF {
				fmt.Println("FillFrom err:", err)
				callback.Disconnected(c, err)
				return
			}

			if ok {
				packetsChan <- &packetChan{
					packet: packet,
					conn:   c,
				}
				i++
				//fmt.Println(i)
				packet = nil
			}

		}
	}
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
