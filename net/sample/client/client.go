package main

import (
	"flag"
	"fmt"
	"os"
	"rapidgo/net"
	"runtime"
	"time"
)

type demoCallback struct {
}

func (callback demoCallback) Disconnected(conn *net.Connection, err error) {
	fmt.Println(conn.RemoteAddr().String(), "disconnected error: ", err)
}

func (callback demoCallback) Connected(conn *net.Connection) {
	//fmt.Println(conn.RemoteAddr().String(), "connected")
}

func (callback demoCallback) Received(conn *net.Connection, packet net.Packet) {
	//fmt.Println(conn.RemoteAddr().String(), "data received")
	//go conn.SendPacket(packet)
}

func main() {
	var ip = flag.String("address", "127.0.0.1:8888", "help message for flagname")
	var num = flag.Int("num", 1000, "connections")
	var size = flag.Int("size", 0xFFFF, "connections")
	flag.Parse()

	runtime.GOMAXPROCS(4)

	if *size > 0xFFFF {
		*size = 0xFFFF
	}
	fmt.Println("address:", *ip)
	fmt.Println("num:", *num)
	fmt.Println("size:", *size)

	for i := 0; i < *num; i++ {
		go func() {
			client := net.CreateTCPClient()
			var demo demoCallback
			//fmt.Println("connecting - ", *ip)
			err := client.Connect(*ip, 100000, demo)
			if err != nil {
				fmt.Println("connect error:", err)
				return
			}

			p := net.DefaultCreatePacketFunc()
			p.Create(uint(*size))

			go func() {
				for {
					select {
					case <-time.Tick(time.Second * 5):
						if err := client.SendPacket(p); err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					}
				}
			}()

			for {
				client.Update()
			}
		}()
		time.Sleep(time.Microsecond * 200)
	}

	select {}
}
