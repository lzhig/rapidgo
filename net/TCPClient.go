package net

import (
	"errors"
	"net"
	"time"
)

// TCPClient type
type TCPClient struct {
	conn Connection // 与服务端的连接

	callback ICallback // 回调接口

	readBuffer []byte // 读取数据缓存

	packetsChan chan *packetChan // 从连接loop的协程传递上主线程
}

// Connect function
func (c *TCPClient) Connect(serverAddress string, timeout uint32, callback ICallback) (err error) {
	if callback == nil {
		return errors.New("Must be set callback")
	}

	c.callback = callback
	c.conn.remoteAddress = serverAddress

	dailer := net.Dialer{
		Timeout:   time.Millisecond * time.Duration(timeout),
		Deadline:  time.Time{},
		KeepAlive: time.Millisecond * time.Duration(30),
	}
	c.conn.conn, err = dailer.Dial("tcp", serverAddress)
	if err != nil {
		return err
	}
	c.callback.Connected(&c.conn)

	c.packetsChan = make(chan *packetChan, 1024)

	go c.conn.loop(callback, c.packetsChan)

	return nil
}

// Disconnect function
func (c *TCPClient) Disconnect() {
	c.conn.disconnect()
}

// Update function
func (c *TCPClient) Update() {
	for p := range c.packetsChan {
		c.callback.Received(p.conn, p.packet)
	}
}

// Send function
func (c *TCPClient) Send(data []byte) error {
	size, err := c.conn.conn.Write(data)
	if err != nil {
		return err
	}
	if size != len(data) {
		return errors.New("Failed to send data")
	}
	return nil
}
