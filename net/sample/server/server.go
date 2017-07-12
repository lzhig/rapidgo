package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"rapidgo/net"
	"runtime"
	"runtime/pprof"
	"syscall"
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

	//go server.SendPacket(conn, packet)

}

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	var ip = flag.String("address", "0.0.0.0:8888", "help message for flagname")
	var num = flag.Int("num", 60000, "connections")
	flag.Parse()
	var demo demoCallback
	fmt.Println("start server - ", *ip)
	err := server.Start(*ip, uint32(*num), demo)
	if err != nil {
		fmt.Println("result: ", err)
		os.Exit(1)
	}
	fmt.Println("started.")

	//go signalListen()

	if err == nil {
		for {
			server.Update()
		}
	}
}

func signalListen() {
	f, e := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	if e != nil {
		log.Fatal(e)
	}
	pprof.StartCPUProfile(f)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	fmt.Println("singal listen")
	for {
		s := <-c
		//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
		fmt.Println("get signal:", s)
		signal.Stop(c)
		pprof.StopCPUProfile()
		f.Close()

		f1, e1 := os.OpenFile("mem.prof", os.O_RDWR|os.O_CREATE, 0644)
		if e1 != nil {
			log.Fatal(e1)
		}
		pprof.WriteHeapProfile(f1)
		f1.Close()

		runtime.GC()

		return
	}
}
