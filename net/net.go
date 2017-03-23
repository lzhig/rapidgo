package net

import (
	"errors"
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

// AtLeastReader interface
type AtLeastReader interface {
	ReadBuffer(buf []byte) bool
}

// CreatePacketFunc type
type CreatePacketFunc func(reader AtLeastReader) (*Packet, error)

var createPacketFunc CreatePacketFunc
var defaultHeader []byte

const (
	defaultHeaderLen = 4
)

func init() {
	createPacketFunc = defaultCreatePacket
	defaultHeader = make([]byte, defaultHeaderLen)
}

// SetCreatePacketFunc function
func SetCreatePacketFunc(f CreatePacketFunc) {
	createPacketFunc = f
}

func defaultCreatePacket(reader AtLeastReader) (*Packet, error) {
	ret := reader.ReadBuffer(defaultHeader)
	if !ret {
		return nil, nil
	}

	if defaultHeader[0] == 0xFE && defaultHeader[1] == 0xDC {
		len := (uint32(defaultHeader[2]) << 8) + uint32(defaultHeader[3])
		p := &Packet{}
		p.init(len)
		return p, nil
	}

	return nil, errors.New("Invalid header")
}

// 用于向上层传递收到的数据包
type packetChan struct {
	packet *Packet
	conn   *Connection
}

// Connection object
type Connection struct {
	remoteAddress string   // 远端地址
	conn          net.Conn // 底层连接

	stopCmdChan  chan bool // 断开时发送此命令
	exitLoopChan chan bool // 当Connection退出loop时，会传入值

	buffer singleReadWriteBuffer // 读取缓存
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

	readBufferLen := uint32(1024 * 1024)
	c.buffer.init(readBufferLen)

	var packet *Packet

	fillFunction := func(p **Packet) {
		io.Copy(*p, &c.buffer)

		if !(*p).canWrite() {
			packetsChan <- &packetChan{
				packet: *p,
				conn:   c,
			}
			*p = nil
		}
	}

	for {
		select {
		case <-c.stopCmdChan:
			c.exitLoopChan <- true
			return
		default:
			// read buffer
			written, err := io.Copy(&c.buffer, c.conn)
			if err != nil {
				callback.Disconnected(c, err)
				return
			}

			if written <= 0 {
				continue
			}

			if packet == nil {
				// 创建packet
				packet, err = createPacketFunc(&c.buffer)
				if err != nil {
					callback.Disconnected(c, err)
				}

				fillFunction(&packet)
			} else {
				fillFunction(&packet)
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
	buf.buffer = make([]byte, buf.len)
}

func (buf *singleReadWriteBuffer) Write(p []byte) (n int, err error) {
	len := uint32(len(p))
	// 没有空余的空间
	if (buf.readPos == 0 && buf.writePos == buf.len-1) || (buf.writePos+1 == buf.readPos) {
		return 0, io.EOF
	}

	num := uint32(0)

	if buf.writePos >= buf.readPos {
		// 如果写位置在读位置右边，先填充从写位置到最右边的空间，如果读位置在0位置处，只能填充到倒数第二个位置
		tailPos := uint32(0)
		readPosIsZero := false
		if buf.readPos == 0 {
			tailPos = buf.len - 1
			readPosIsZero = true
		} else {
			tailPos = buf.len
		}

		num = uint32(copy(buf.buffer[buf.writePos:tailPos], p))
		buf.writePos += num

		if num == len || readPosIsZero {
			return int(num), nil
		}

		// 如果还没有完成, 第二次填充最多只能到读位置前一个位置
		buf.writePos = 0
	}

	tailPos := buf.readPos - 1
	num2 := uint32(copy(buf.buffer[buf.writePos:tailPos], p[num:]))
	buf.writePos += num2
	return int(num + num2), nil
}

func (buf *singleReadWriteBuffer) Read(p []byte) (n int, err error) {
	if buf.readPos == buf.writePos {
		return 0, nil
	}

	if buf.readPos < buf.writePos {
		num := copy(p, buf.buffer[buf.readPos:buf.writePos])
		buf.readPos += uint32(num)
		return num, nil
	}

	num := copy(p, buf.buffer[buf.readPos:])
	buf.readPos += uint32(num)

	if buf.readPos < buf.len {
		return num, nil
	}

	buf.readPos = 0

	if buf.writePos == 0 {
		return num, nil
	}

	num2 := copy(p[num:], buf.buffer[:buf.writePos+1])
	buf.readPos += uint32(num2)

	return num + num2, nil
}

func (buf *singleReadWriteBuffer) ReadBuffer(p []byte) bool {
	readLen := uint32(len(p))

	if buf.readPos < buf.writePos {
		if buf.readPos+readLen > buf.writePos {
			return false
		}

		copy(p, buf.buffer[buf.readPos:buf.readPos+readLen])
		buf.readPos += readLen

		return true
	}

	if buf.readPos+readLen-buf.len <= buf.writePos {
		copy(p, buf.buffer[buf.readPos:buf.len])
		copy(p[buf.len-buf.readPos:], buf.buffer[:buf.readPos+readLen-buf.len])
		buf.readPos = buf.readPos + readLen - buf.len
		return true
	}

	return false
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
