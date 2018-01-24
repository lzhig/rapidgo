package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	_ "net/http/pprof"

	"github.com/lzhig/rapidgo/net"
)

var server = net.CreateTCPServer()

func main() {
	runtime.GOMAXPROCS(4)
	go func() {
		http.ListenAndServe("0.0.0.0:8092", nil)
	}()

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
				fmt.Println(event.Conn.RemoteAddr().String(), "disconnected")
			}
		}
	}
}

func handleConnection(conn *net.Connection) {
	for {
		select {
		case data := <-conn.DataChan:
			if data == nil {
				return
			}
			conn.Send(data)
		}
	}
}
