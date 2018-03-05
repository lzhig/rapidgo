package rapidnet

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// 用于向上层传递收到的数据包
type dataChan struct {
	data []byte
	conn *Connection
}

// Connection object
type Connection struct {
	remoteAddress string   // 远端地址
	conn          net.Conn // 底层连接

	packetHandler PacketHandler // 包处理器

	receiveDataChan chan []byte
	sendDataChan    chan []byte

	stopCmdChan      chan struct{} // 断开时发送此命令
	stopSendLoopChan chan struct{}

	release func()

	releaseOnce sync.Once
}

func (c *Connection) init() {
	c.receiveDataChan = make(chan []byte, 16)
	c.sendDataChan = make(chan []byte, 16)
	c.stopCmdChan = make(chan struct{})
	c.stopSendLoopChan = make(chan struct{})
}

// ReceiveDataChan 返回连接接收到的数据chan
func (c *Connection) ReceiveDataChan() <-chan []byte {
	return c.receiveDataChan
}

// RemoteAddr function
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Disconnect() {
	fmt.Println("call Connonection Disconnect()")
	c.releaseOnce.Do(c._disconnect)
}

func (c *Connection) _disconnect() {
	c.conn.Close()
	close(c.stopCmdChan)
}

func (c *Connection) loop(eventChan chan *Event) {
	defer c.conn.Close()

	go c.sendLoop(eventChan)
	defer close(c.receiveDataChan)
	defer close(c.stopSendLoopChan)
	defer c.release()

	for {
		select {
		case <-c.stopCmdChan:
			eventChan <- &Event{Type: EventDisconnected, Err: errors.New("stopped"), Conn: c}
			return

		default:
			data, err := c.packetHandler.Receive()
			if err != nil {
				eventChan <- &Event{Type: EventDisconnected, Err: err, Conn: c}
				return
			}

			if data != nil {
				c.receiveDataChan <- data
			}
		}
	}
}

func (c *Connection) sendLoop(eventChan chan *Event) {
	for {
		select {
		case <-c.stopSendLoopChan:
			return

		case data := <-c.sendDataChan:
			if err := c.packetHandler.Send(data); err != nil {
				eventChan <- &Event{Type: EventSendFailed, Err: err, Conn: c}
				return
			}
		}
	}
}

// Send send data
func (c *Connection) Send(data []byte) {
	select {
	case c.sendDataChan <- data:
	default:
		panic(errors.New("[rapidnet] connection Send: sendDataChan is full"))
	}
}

// 管理建立的连接
type connections struct {
	connections map[*Connection]*Connection
	mutex       sync.Mutex
	sem         chan struct{}
}

func (conns *connections) init(n uint32) {
	conns.connections = make(map[*Connection]*Connection, n)
	conns.sem = make(chan struct{}, n)
}

func (conns *connections) size() uint32 {
	return uint32(len(conns.connections))
}

func (conns *connections) add(conn *Connection) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()

	conns.connections[conn] = conn
}

func (conns *connections) remove(conn *Connection) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()

	delete(conns.connections, conn)
	conns.release()
}

func (conns *connections) acquire() { conns.sem <- struct{}{} }
func (conns *connections) release() { <-conns.sem }

// CreateTCPClient creates a client object for tcp
func CreateTCPClient() *TCPClient {
	return &TCPClient{}
}

// CreateTCPServer function a server object for tcp
func CreateTCPServer() *TCPServer {
	return &TCPServer{}
}
