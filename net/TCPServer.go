package net

import (
	"net"

	"errors"
)

// TCPServer struct
type TCPServer struct {
	stopCmdChan  chan bool
	exitLoopChan chan bool

	maxClientsCount uint32
	conns           connections

	callback ICallback

	packetsChan chan *packetChan
}

// Start function
func (s *TCPServer) Start(address string, maxClientsAllowed uint32, callback ICallback) (err error) {
	if callback == nil {
		return errors.New("Must be set callback")
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}

	netListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	s.callback = callback
	s.stopCmdChan = make(chan bool, 1)
	s.exitLoopChan = make(chan bool, 1)
	s.packetsChan = make(chan *packetChan, 4096)

	s.maxClientsCount = maxClientsAllowed
	s.conns.init(maxClientsAllowed)

	go s.loop(netListener)

	return nil
}

// Stop function
func (s *TCPServer) Stop() {
	s.stopCmdChan <- true
	<-s.exitLoopChan
}

func (s *TCPServer) loop(netListener *net.TCPListener) {
	defer netListener.Close()

	for {
		select {
		case <-s.stopCmdChan:
			s.exitLoopChan <- true
			return
		default:
			//netListener.SetDeadline(time.Now().Add(time.Millisecond))
			s.conns.acquire()

			conn, err := netListener.AcceptTCP()
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				s.conns.release()
				continue
			} else if err != nil {
				s.conns.release()
				return
			}

			newConn := &Connection{conn: conn}
			s.conns.add(newConn)

			s.callback.Connected(newConn)

			go newConn.loop(s, s.packetsChan)
		}
	}
}

// Disconnected function
func (s *TCPServer) Disconnected(conn *Connection, err error) {
	s.callback.Disconnected(conn, err)
	s.conns.remove(conn)
}

// Connected function
func (s *TCPServer) Connected(conn *Connection) {
	s.callback.Connected(conn)
}

// Received function
func (s *TCPServer) Received(conn *Connection, packet Packet) {
	s.callback.Received(conn, packet)
}

// Update function
func (s *TCPServer) Update() {
	for p := range s.packetsChan {
		s.callback.Received(p.conn, p.packet)
	}
}

// Send function
func (s *TCPServer) Send(conn *Connection, data []byte) error {
	return conn.Send(data)
}

func (s *TCPServer) SendPacket(conn *Connection, p Packet) error {
	return conn.SendPacket(p)
}
