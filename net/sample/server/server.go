package main

import "rapidgo/net"
import "fmt"

type demoCallback struct {
}

func (callback demoCallback) Disconnected(conn *net.Connection, err error) {
	fmt.Println(conn.RemoteAddr().String())
	fmt.Println("disconnected error: ", err)
}

func (callback demoCallback) Connected(conn *net.Connection) {
	fmt.Println(conn.RemoteAddr().String())
	fmt.Println("connected")
}

func (callback demoCallback) Received(conn *net.Connection, packet *net.Packet) {
	fmt.Println(conn.RemoteAddr().String())
	fmt.Println("data received.")
}

func main() {
	server := net.CreateTCPServer()
	var demo demoCallback
	fmt.Println("start server...")
	err := server.Start("0.0.0.0:8888", 1, demo)
	fmt.Println("result: ", err)

	if err == nil {
		select {}
	}
}
