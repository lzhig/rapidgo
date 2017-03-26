package main

import (
	"fmt"
	"os"
	"rapidgo/net"
	"runtime"
)

type demoCallback struct {
}

func (callback demoCallback) Disconnected(conn *net.Connection, err error) {
	fmt.Println(conn.RemoteAddr().String(), "disconnected error: ", err)
}

func (callback demoCallback) Connected(conn *net.Connection) {
	fmt.Println(conn.RemoteAddr().String(), "connected")
}

var server = net.CreateTCPServer()

func (callback demoCallback) Received(conn *net.Connection, packet net.Packet) {
	fmt.Println(conn.RemoteAddr().String(), "data received")

	buf := packet.GetBuffer()
	len := len(buf)
	header := make([]byte, 4)
	header[0] = 0xfe
	header[1] = 0xdc
	header[2] = byte(len >> 8)
	header[3] = byte(len & 0xff)
	server.Send(conn, header)
	server.Send(conn, packet.GetBuffer())
}

func main() {
	runtime.GOMAXPROCS(4)
	var demo demoCallback
	fmt.Println("start server...")
	err := server.Start("127.0.0.1:8888", 2000, demo)
	if err != nil {
		fmt.Println("result: ", err)
		os.Exit(1)
	}
	fmt.Println("started.")

	if err == nil {
		for {
			server.Update()
		}
	}
}
