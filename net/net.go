package net

import (
	"io"
	"net"
	"sync"
)

// ICallback interface
type ICallback interface {
	Disconnected(conn *Connection, err error)
	Connected(conn *Connection)
	Received(conn *Connection, packet *Packet)
}

// Connection object
type Connection struct {
	remoteAddress string   // 远端地址
	conn          net.Conn // 底层连接

	stopCmdChan  chan bool // 断开时发送此命令
	exitLoopChan chan bool // 当Connection退出loop时，会传入值

	buffer singleReadWriteBuffer // 读取缓存
}

// 用于向上层传递收到的数据包
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

type AtLeastReader interface {
	ReadAtLeast(r io.Reader, buf []byte, min int) (n int, err error)
}

type CreatePacketFunc func(reader AtLeastReader) *Packet

func defaultCreatePacket(reader AtLeastReader) *Packet {
	return nil
}

func SetCreatePacketFunc(f CreatePacketFunc) {
	createPacketFunc = f
}

var createPacketFunc CreatePacketFunc

func init() {
	createPacketFunc = defaultCreatePacket
}

func (c *Connection) loop(callback ICallback, packetsChan chan *packetChan) {
	defer c.conn.Close()

	readBufferLen := uint32(1024 * 1024)
	c.buffer.init(readBufferLen)

	var packet *Packet

	for {
		select {
		case <-c.stopCmdChan:
			c.exitLoopChan <- true
			return
		default:
			// read buffer
			b := c.buffer.getWriteBuffer()
			if len(b) <= 0 {
				panic("the len of writebuffer should not be zero.")
			}

			n, err := c.conn.Read(b)
			if err != nil {
				callback.Disconnected(c, err)
				return
			}

			if n <= 0 {
				continue
			}
			c.buffer.moveWritePos(uint32(n))

			if packet == nil {
				// 创建packet
				packet = createPacketFunc(c.buffer)
			}
		}
	}
}

// 单向循环读写缓冲
type singleReadWriteBuffer struct {
	buffer   []byte // 缓冲区
	readPos  uint32 // 读位置
	writePos uint32 // 写位置
	len      uint32 // 缓冲区大小
}

func (buf *singleReadWriteBuffer) init(len uint32) {
	buf.len = len
	buf.buffer = make([]byte, len)
}

func (buf *singleReadWriteBuffer) getWriteBuffer() []byte {
	if buf.writePos >= buf.readPos {
		return buf.buffer[buf.writePos:]
	}
	return buf.buffer[buf.writePos : buf.readPos-1]
}

func (buf *singleReadWriteBuffer) moveWritePos(pos uint32) {
	p := buf.writePos + pos

	if buf.writePos >= buf.readPos {
		// 当写位置大于读位置时，如果移动的位置不超过总长度
		if p < buf.len {
			buf.writePos = p
			return
		}

		// 否则这个位置又移动头部，但不是超过读位置
		p = p - buf.len
		if p > buf.readPos {
			panic("pos too large.")
		}
		buf.writePos = p
		return
	}

	// 当写位置小于读位置时，移动位置不能超过读位置
	if p >= buf.readPos {
		panic("pos too large.")
	}
	buf.writePos = p
}

func (buf *singleReadWriteBuffer) ReadAtLeast(r io.Reader, min int) (n int, err error) {
	readLen = len(p)
}

// 管理建立的连接
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

// CreateTCPClient creates a client object for tcp
func CreateTCPClient() *TCPClient {
	return &TCPClient{}
}

// CreateTCPServer function a server object for tcp
func CreateTCPServer() *TCPServer {
	return &TCPServer{}
}
