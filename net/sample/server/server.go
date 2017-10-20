package main

import (
	"flag"
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
	//fmt.Println(conn.RemoteAddr().String(), "data received")

	server.SendPacket(conn, packet)
}

func main() {
	runtime.GOMAXPROCS(4)
	var ip = flag.String("address", "0.0.0.0:8888", "help message for flagname")
	var num = flag.Int("num", 10000, "connections")
	flag.Parse()
	var demo demoCallback
	fmt.Println("start server - ", *ip)
	err := server.Start(*ip, uint32(*num), demo)
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
