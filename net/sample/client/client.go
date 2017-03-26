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

func (callback demoCallback) Received(conn *net.Connection, packet net.Packet) {
	fmt.Println(conn.RemoteAddr().String(), "data received")
}

func main() {
	runtime.GOMAXPROCS(4)
	for i := 0; i < 1; i++ {
		go func() {
			client := net.CreateTCPClient()
			var demo demoCallback
			fmt.Println("connecting...")
			err := client.Connect("127.0.0.1:8888", 5000, demo)
			if err != nil {
				fmt.Println("connect error:", err)
				return
			}

			go func() {
				len := 65535
				d := make([]byte, len+4)
				d[0] = 0xfe
				d[1] = 0xdc
				d[2] = byte(len >> 8)
				d[3] = byte(len & 0xff)

				for {
					if err := client.Send(d); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}()

			for {
				client.Update()
			}
		}()
	}
	select {}
}
