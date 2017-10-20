package net

import "io"
import "errors"

// Packet interface
type Packet interface {
	Create(size uint)
	FillFrom(r io.Reader) (ok bool, err error)
	Read(data []byte) (uint32, error)
	GetBuffer() []byte
	Send(c *Connection) error
}

// DefaultCreatePacketFunc is default function to create packet
var DefaultCreatePacketFunc = func() Packet {
	return &defaultPacket{}
}

const (
	defaultHeaderSize = 4
)

type defaultPacket struct {
	header [defaultHeaderSize]byte // 头部数据

	buffer []byte // 缓冲区

	readPos  uint // 读位置
	writePos uint // 写位置
	len      uint // 缓冲区大小

	filledHeaderSize uint8 // 已经填充了多少头部数据
}

func (p *defaultPacket) init(len uint) {
	p.buffer = make([]byte, len)
	p.len = len
}

func (p *defaultPacket) Write(buf []byte) (n int, err error) {
	num := copy(p.buffer[p.writePos:], buf)
	p.writePos += uint(num)
	return num, nil
}

func (p *defaultPacket) FillFrom(r io.Reader) (ok bool, err error) {
	if p.filledHeaderSize < defaultHeaderSize {
		n, err := r.Read(p.header[p.filledHeaderSize:])
		p.filledHeaderSize += uint8(n)
		if err != nil {
			return false, err
		}
		if p.filledHeaderSize < defaultHeaderSize {
			return false, nil
		}

		// 判断标志位
		if p.header[0] != 0xFE || p.header[1] != 0xDC {
			return false, errors.New("invalid connection")
		}

		p.len = (uint(p.header[2]) << 8) + uint(p.header[3])
		p.buffer = make([]byte, p.len)
	}

	n, err := r.Read(p.buffer[p.writePos:])
	p.writePos += uint(n)
	if err != nil {
		return false, err
	}

	if p.writePos == p.len {
		return true, nil
	}

	return false, nil
}

func (p *defaultPacket) Read(data []byte) (uint32, error) {
	n := copy(data, p.buffer[p.readPos:])
	return uint32(n), nil
}

func (p *defaultPacket) GetBuffer() []byte {
	return p.buffer
}

func (p *defaultPacket) Send(c *Connection) error {
	if err := c.Send(p.header[:]); err != nil {
		return err
	}
	if err := c.Send(p.buffer); err != nil {
		return err
	}
	return nil
}

func (p *defaultPacket) Create(size uint) {
	p.header[0] = 0xFE
	p.header[1] = 0xDC
	p.header[2] = byte((size & 0xFF00) >> 8)
	p.header[3] = byte(size & 0xFF)
	p.init(size)
}
