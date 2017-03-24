package net

import "io"
import "errors"

// Packet interface
type Packet interface {
	FillFrom(r io.Reader) (ok bool, err error)
}

// PacketFactory interface
type PacketFactory interface {
	CreatePacket() Packet
}

var packetFactory PacketFactory

func init() {
	packetFactory = &defaultPacketFactory{}
}

// SetPacketFactory function
func SetPacketFactory(f PacketFactory) {
	packetFactory = f
}

type defaultPacketFactory struct {
}

func (f *defaultPacketFactory) CreatePacket() Packet {
	return &defaultPacket{}
}

const (
	defaultHeaderSize = 4
)

type defaultPacket struct {
	header [defaultHeaderSize]byte // 头部数据

	buffer []byte // 缓冲区

	readPos  uint32 // 读位置
	writePos uint32 // 写位置
	len      uint32 // 缓冲区大小

	filledHeaderSize uint8 // 已经填充了多少头部数据
}

func (p *defaultPacket) init(len uint32) {
	p.buffer = make([]byte, len)
	p.len = len
}

func (p *defaultPacket) Write(buf []byte) (n int, err error) {
	num := copy(p.buffer[p.writePos:], buf)
	p.writePos += uint32(num)
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

		p.len = (uint32(p.header[2]) << 8) + uint32(p.header[3])
		p.buffer = make([]byte, p.len)
	}

	n, err := r.Read(p.buffer[p.writePos:])
	p.writePos += uint32(n)
	if err != nil {
		return false, err
	}

	if p.writePos == p.len {
		return true, nil
	}

	return false, nil
}
