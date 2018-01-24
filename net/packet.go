package net

import (
	"bufio"
	"errors"
	"net"
	"time"
)

type PacketHandler interface {
	Receive() ([]byte, error)
	Send([]byte) error
}

// // Packet interface
// type Packet interface {
// 	Create(size uint)
// 	FillFrom(r io.Reader) (ok bool, err error)
// 	Read(data []byte) (uint32, error)
// 	GetBuffer() []byte
// 	Send(c *Connection) error
// }

const (
	defaultTag        = 0xDCFE
	defaultTagSize    = 2 // 2bytes
	defaultLenSize    = 2 //2 bytes
	defaultHeaderSize = defaultTagSize + defaultLenSize
)

type defaultPacketHandler struct {
	conn      net.Conn
	bufReader *bufio.Reader
	bufWriter *bufio.Writer

	data        []byte
	dataLen     int
	readed      int
	headerReady bool
}

func (obj *defaultPacketHandler) Receive() ([]byte, error) {
	if !obj.headerReady {
		// 读取header
		obj.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		p, err := obj.bufReader.Peek(defaultHeaderSize)
		if err == nil && len(p) == defaultHeaderSize {
			if p[0] != 0xFE || p[1] != 0xDC {
				return nil, errors.New("invalid data")
			}

			obj.bufReader.Discard(defaultHeaderSize)
			obj.headerReady = true

			obj.dataLen = (int(p[2]) << 8) + int(p[3])
			obj.data = make([]byte, obj.dataLen)
		} else {
			nerr, ok := err.(net.Error)
			if ok {
				if nerr.Timeout() {
					return nil, nil
				} else if nerr.Temporary() {
					return nil, nerr
				}
			}
		}
	}

	// read body

	obj.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	n, err := obj.bufReader.Read(obj.data[obj.readed:])
	if err == nil {
		obj.readed += n
		if obj.readed == obj.dataLen {
			p := obj.data
			obj.data = nil
			obj.readed = 0
			obj.headerReady = false
			obj.dataLen = 0
			return p, nil
		}
	}

	return nil, nil
}

func (obj *defaultPacketHandler) Send(data []byte) error {
	if len(data) > 0xFFFF {
		return errors.New("too large")
	}

	n, err := obj.bufWriter.Write([]byte{0xFE, 0xDC})
	if n != 2 || err != nil {
		obj.conn.Close()
		return err
	}

	len := len(data)

	err = obj.bufWriter.WriteByte(byte((len & 0xFF00) >> 8))
	if err != nil {
		obj.conn.Close()
		return err
	}

	err = obj.bufWriter.WriteByte(byte(len & 0xFF))
	if err != nil {
		obj.conn.Close()
		return err
	}

	n, err = obj.bufWriter.Write(data)
	if n != len || err != nil {
		obj.conn.Close()
		return err
	}
	obj.bufWriter.Flush()
	return nil
}

// type defaultPacket struct {
// 	//header [defaultHeaderSize]byte // 头部数据
// 	header []byte

// 	buffer []byte // 缓冲区

// 	readPos  uint // 读位置
// 	writePos uint // 写位置
// 	len      uint // 缓冲区大小

// 	filledHeaderSize uint8 // 已经填充了多少头部数据
// }

// func (p *defaultPacket) init(dataLen uint) {
// 	p.buffer = make([]byte, dataLen+defaultHeaderSize)
// 	p.len = dataLen + defaultHeaderSize
// }

// func (p *defaultPacket) Write(buf []byte) int {
// 	num := copy(p.buffer[p.writePos:], buf)
// 	p.writePos += uint(num)
// 	return num
// }

// func (p *defaultPacket) FillFrom(r io.Reader) (ok bool, err error) {
// 	if p.filledHeaderSize < defaultHeaderSize {
// 		n, err := r.Read(p.header[p.filledHeaderSize:])
// 		p.filledHeaderSize += uint8(n)
// 		if err != nil {
// 			return false, err
// 		}
// 		if p.filledHeaderSize < defaultHeaderSize {
// 			return false, nil
// 		}

// 		// 判断标志位
// 		if p.header[0] != 0xFE || p.header[1] != 0xDC {
// 			return false, errors.New("invalid connection")
// 		}

// 		p.len = (uint(p.header[2]) << 8) + uint(p.header[3])
// 		p.buffer = make([]byte, p.len)
// 	}

// 	n, err := r.Read(p.buffer[p.writePos:])
// 	p.writePos += uint(n)
// 	if err != nil {
// 		return false, err
// 	}

// 	if p.writePos == p.len {
// 		return true, nil
// 	}

// 	return false, nil
// }

// func (p *defaultPacket) Read(data []byte) (uint32, error) {
// 	n := copy(data, p.buffer[p.readPos:])
// 	return uint32(n), nil
// }

// func (p *defaultPacket) GetBuffer() []byte {
// 	return p.buffer
// }

// func (p *defaultPacket) Send(c *Connection) error {
// 	if err := c.Send(p.header[:]); err != nil {
// 		return err
// 	}
// 	if err := c.Send(p.buffer); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (p *defaultPacket) Create(size uint) {
// 	p.header[0] = 0xFE
// 	p.header[1] = 0xDC
// 	p.header[2] = byte((size & 0xFF00) >> 8)
// 	p.header[3] = byte(size & 0xFF)
// 	p.init(size)
// }
