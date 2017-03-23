package net

// Packet struct
type Packet struct {
	buffer   []byte // 缓冲区
	readPos  uint32 // 读位置
	writePos uint32 // 写位置
	len      uint32 // 缓冲区大小
}

func (p *Packet) init(len uint32) {
	p.buffer = make([]byte, len)
	p.len = len
}

func (p *Packet) Write(buf []byte) (n int, err error) {
	num := copy(p.buffer[p.writePos:], buf)
	p.writePos += uint32(num)
	return num, nil
}

func (p *Packet) canWrite() bool {
	return p.writePos < p.len
}
