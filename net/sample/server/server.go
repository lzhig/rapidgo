package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

	origin "net"

	"github.com/lzhig/rapidgo/net"
)

const (
	defaultTag        = 0xDCFE
	defaultTagSize    = 2 // 2bytes
	defaultLenSize    = 4 //2 bytes
	defaultHeaderSize = defaultTagSize + defaultLenSize
)

type packetHandler struct {
	conn      origin.Conn
	bufReader *bufio.Reader
	bufWriter *bufio.Writer

	data        []byte
	dataLen     int
	readed      int
	headerReady bool
}

func (obj *packetHandler) Receive() ([]byte, error) {
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

			obj.dataLen = (int(p[5]) << 24) + (int(p[4]) << 16) + (int(p[3]) << 8) + int(p[2])
			obj.data = make([]byte, obj.dataLen)
		} else {
			nerr, ok := err.(origin.Error)
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

func (obj *packetHandler) Send(data []byte) error {
	len := len(data)

	if len > 0x7FFFFFFF {
		return errors.New("too large")
	}

	n, err := obj.bufWriter.Write([]byte{0xFE, 0xDC, byte(len & 0xFF), byte((len & 0xFF00) >> 8), byte((len & 0xFF0000) >> 16), byte((len & 0xFF000000) >> 24)})
	if n != defaultHeaderSize || err != nil {
		obj.conn.Close()
		return err
	}

	n, err = obj.bufWriter.Write(data)
	if n != len || err != nil {
		obj.conn.Close()
		return err
	}

	return nil
}

var server = net.CreateTCPServer()

func main() {
	runtime.GOMAXPROCS(4)
	go func() {
		http.ListenAndServe("0.0.0.0:8092", nil)
	}()

	config := &net.Config{PacketHandlerFactory: func(c origin.Conn) net.PacketHandler {
		return &packetHandler{
			conn:      c,
			bufReader: bufio.NewReader(c),
			bufWriter: bufio.NewWriter(c),
		}
	},
	}
	net.Init(config)

	var ip = flag.String("address", "0.0.0.0:8888", "help message for flagname")
	var num = flag.Int("num", 10000, "connections")
	flag.Parse()
	fmt.Println("start server - ", *ip)
	eventChan, err := server.Start(*ip, uint32(*num))
	if err != nil {
		fmt.Println("result: ", err)
		os.Exit(1)
	}
	fmt.Println("started.")

	for {
		select {
		case event := <-eventChan:
			switch event.Type {
			case net.EventConnected:
				fmt.Println(event.Conn.RemoteAddr().String(), "connected")
				go handleConnection(event.Conn)
			case net.EventDisconnected:
				fmt.Println(event.Conn.RemoteAddr().String(), "disconnected", event.Err)

			case net.EventSendFailed:
				fmt.Println(event.Conn.RemoteAddr().String(), "Failed to send", event.Err)
			}
		}
	}
}

func handleConnection(conn *net.Connection) {
	for {
		select {
		case data := <-conn.ReceiveDataChan():
			if data == nil {
				return
			}
			fmt.Println("Recieve data. size:", len(data))
			conn.Send(data)
		}
	}
}
