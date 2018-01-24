package network

import (
	"errors"
	"net"
	"sync"
)

// ErrServerClosed is returned by the Server's Serve, ServeTLS, ListenAndServe,
// and ListenAndServeTLS methods after a call to Shutdown or Close.
var ErrServerClosed = errors.New("TCPServer: Server closed")

type TCPServerConfig struct {
	MaxConnections uint // 最大连接数
	KeepAlive      bool
}

// TCPServer type
type TCPServer struct {
	config *TCPServerConfig // 配置

	mu       sync.Mutex
	doneChan chan struct{}
}

// Start method
func (obj *TCPServer) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return obj.Serve(ln)
}

// Serve method
func (obj *TCPServer) Serve(l net.Listener) error {
	defer l.Close()

	for {
		conn, e := l.Accept()
		if e != nil {
			select {
			case <-obj.getDoneChan():
				return ErrServerClosed
			default:
			}
			return e
		}

		c := obj.newConn(conn)
		go c.serve()
	}
}

// Stop method
func (obj *TCPServer) Stop() {

}

func (obj *TCPServer) getDoneChan() <-chan struct{} {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	return obj.getDoneChanLocked()
}

func (obj *TCPServer) getDoneChanLocked() chan struct{} {
	if obj.doneChan == nil {
		obj.doneChan = make(chan struct{})
	}
	return obj.doneChan
}

func (obj *TCPServer) closeDoneChanLocked() {
	ch := obj.getDoneChanLocked()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by s.mu.
		close(ch)
	}
}

// Create new connection from rwc.
func (obj *TCPServer) newConn(rwc net.Conn) *TCPConnection {
	c := &TCPConnection{
		server: obj,
		rwc:    rwc,
	}
	// if debugServerConnections {
	// 	c.rwc = newLoggingConn("server", c.rwc)
	// }
	return c
}
