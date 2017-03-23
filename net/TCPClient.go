package net

import (
	"errors"
	"net"
	"time"
)

// TCPClient type
type TCPClient struct {
	conn Connection

	callback Callback

	readBuffer []byte

	packetsChan chan *packetChan
}

// Connect function
func (c *TCPClient) Connect(serverAddress string, timeout uint32, callback Callback) (err error) {
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
