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

func (callback demoCallback) Received(conn *net.Connection, packet net.Packet) {
	//fmt.Println(conn.RemoteAddr().String(), "data received")
	conn.SendPacket(packet)
}

func main() {
	var ip = flag.String("address", "127.0.0.1:8888", "help message for flagname")
	var num = flag.Int("num", 1000, "connections")
	flag.Parse()
	runtime.GOMAXPROCS(4)
	for i := 0; i < *num; i++ {
		go func() {
			client := net.CreateTCPClient()
			var demo demoCallback
			fmt.Println("connecting - ", *ip)
			err := client.Connect(*ip, 5000000, demo)
			if err != nil {
				fmt.Println("connect error:", err)
				return
			}

			p := net.DefaultCreatePacketFunc()
			p.Create(65535)

			if err := client.SendPacket(p); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for {
				client.Update()
			}
		}()
	}
	select {}
}
