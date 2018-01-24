package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/lzhig/rapidgo/net"
)

func main() {
	var ip = flag.String("address", "192.168.2.50:8888", "help message for flagname")
	var num = flag.Int("num", 1, "connections")
	flag.Parse()
	runtime.GOMAXPROCS(4)
	for i := 0; i < *num; i++ {
		go func() {
			client := net.CreateTCPClient()
			fmt.Println("connecting - ", *ip)
			conn, eventChan, err := client.Connect(*ip, 5000000)
			if err != nil {
				fmt.Println("connect error:", err)
				return
			}

			p := make([]byte, 65535, 65535)

			if err := conn.Send(p); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("send ok")

			for {
				select {
				case event := <-eventChan:
					switch event.Type {
					case net.EventConnected:
						fmt.Println(event.Conn.RemoteAddr().String(), "connected")
					case net.EventDisconnected:
						fmt.Println(event.Conn.RemoteAddr().String(), "disconnected.", event.Err)
						return
					}

				case data := <-conn.DataChan:
					conn.Send(data)
				}
			}
		}()
	}
	select {}
}
