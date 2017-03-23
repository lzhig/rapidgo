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
	client := net.CreateTCPClient()
	var demo demoCallback
	fmt.Println("connecting...")
	err := client.Connect("127.0.0.1:8888", 5000, demo)
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}

	for {
		client.Update()
	}
}
