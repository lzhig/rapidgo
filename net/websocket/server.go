package websocket

import (
	"log"
	"net"
	"net/http"
	ws "rapidgo/net/websocket/websocket"
	"sync"
)

type ICallback interface {
	Disconnected(conn *Connection, err error)
	Connected(conn *Connection)
	Received(conn *Connection, data []byte)
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type handler struct {
	server   *Server
	pattern  string
	callback ICallback
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	Connection := &Connection{conn: conn}
	Connection.stopCmdChan = make(chan bool, 1)
	Connection.exitLoopChan = make(chan bool, 1)

	h.callback.Connected(Connection)

	go Connection.writeloop()

	Connection.readloop(h.server.packetChan, h.callback)
}

type packet struct {
	conn     *Connection
	callback ICallback
	data     []byte
}

// Server class
type Server struct {
	CheckOrigin bool

	packetChan chan *packet

	stopCmdChan  chan bool
	exitLoopChan chan bool
}

// Start Function: start websocket server
func (s *Server) Start(addr string, maxConns uint) error {
	if !s.CheckOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	l = LimitListener(l, maxConns)

	s.stopCmdChan = make(chan bool, 1)
	s.exitLoopChan = make(chan bool, 1)
	s.packetChan = make(chan *packet, 4096)

	return http.Serve(l, nil)
}

// Stop function
func (s *Server) Stop() {
	s.stopCmdChan <- true
	<-s.exitLoopChan
}

// Register -
func (s *Server) Register(pattern string, callback ICallback) {
	http.Handle(pattern, &handler{server: s, pattern: pattern, callback: callback})
}

func (s *Server) Update() {
	select {
	case p := <-s.packetChan:
		p.callback.Received(p.conn, p.data)
	default:
	}
}

// Connection type
type Connection struct {
	conn *ws.Conn

	stopCmdChan  chan bool // 断开时发送此命令
	exitLoopChan chan bool // 当Connection退出loop时，会传入值
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() {
	c.stopCmdChan <- true
	<-c.exitLoopChan
	c.conn.Close()
}

func (c *Connection) Send(data []byte) error {
	return c.conn.WriteMessage(ws.BinaryMessage, data)
}

func (c *Connection) readloop(packetChan chan *packet, callback ICallback) {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case <-c.stopCmdChan:
			c.exitLoopChan <- true
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				//if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway) {
				//	log.Printf("error: %v", err)
				//}
				//break
				callback.Disconnected(c, err)
				return
			}
			packetChan <- &packet{conn: c, callback: callback, data: message}
			//fmt.Println("new message")
		}
	}
}

func (c *Connection) writeloop() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case <-c.stopCmdChan:
			c.exitLoopChan <- true
			return
		default:
		}
	}
}

// LimitListener returns a Listener that accepts at most n simultaneous
// connections from the provided Listener.
func LimitListener(l net.Listener, n uint) net.Listener {
	return &limitListener{l, make(chan struct{}, n)}
}

type limitListener struct {
	net.Listener
	sem chan struct{}
}

func (l *limitListener) acquire() { l.sem <- struct{}{} }
func (l *limitListener) release() { <-l.sem }

func (l *limitListener) Accept() (net.Conn, error) {
	//若connect chan已满,则会阻塞在此处
	l.acquire()
	c, err := l.Listener.Accept()
	if err != nil {
		l.release()
		return nil, err
	}
	return &limitListenerConn{Conn: c, release: l.release}, nil
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}
