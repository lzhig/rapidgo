package net

import (
	"net"
	"time"
)

// TCPClient type
type TCPClient struct {
	conn Connection // 与服务端的连接

	readBuffer []byte // 读取数据缓存

	eventChan chan *Event
}

// Connect function
func (c *TCPClient) Connect(serverAddress string, timeout uint32) (*Connection, <-chan *Event, error) {
	c.conn.remoteAddress = serverAddress

	dailer := net.Dialer{
		Timeout:   time.Millisecond * time.Duration(timeout),
		Deadline:  time.Time{},
		KeepAlive: time.Second * time.Duration(30),
	}
	var err error
	c.conn.conn, err = dailer.Dial("tcp", serverAddress)
	if err != nil {
		return nil, nil, err
	}
	c.eventChan = make(chan *Event, 2)
	c.conn.DataChan = make(chan []byte, 16)
	c.eventChan <- &Event{Type: EventConnected, Conn: &c.conn}

	c.conn.packetHandler = config.PacketHandlerFactory(c.conn.conn)

	go c.conn.loop(c.eventChan)

	return &c.conn, c.eventChan, nil
}

// Disconnect function
func (c *TCPClient) Disconnect() {
	c.conn.disconnect()
}
