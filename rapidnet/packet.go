package rapidnet

import (
	"bufio"
	"errors"
	"io"
	"net"
	"time"
)

type PacketHandler interface {
	Receive() ([]byte, error)
	Send([]byte) error
}

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
		if err == io.EOF {
			return nil, err
		}
		if err == nil && len(p) == defaultHeaderSize {
			if p[0] != 0xFE || p[1] != 0xDC {
				return nil, errors.New("invalid data")
			}

			obj.bufReader.Discard(defaultHeaderSize)
			obj.headerReady = true

			obj.dataLen = int(p[2]) + int(p[3]<<8)
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
	} else if err == io.EOF {
		return nil, err
	}

	return nil, nil
}

func (obj *defaultPacketHandler) Send(data []byte) error {
	len := len(data)

	if len > 0xFFFF {
		return errors.New("too large")
	}

	n, err := obj.bufWriter.Write([]byte{0xFE, 0xDC, byte(len & 0xFF), byte((len & 0xFF00) >> 8)})
	if n != 4 || err != nil {
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
