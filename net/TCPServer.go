package net

import (
	"net"

	"errors"
)

// TCPServer struct
type TCPServer struct {
	netListener *net.TCPListener

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

	s.netListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	s.callback = callback
	s.stopCmdChan = make(chan bool, 1)
	s.exitLoopChan = make(chan bool, 1)
	s.packetsChan = make(chan *packetChan, 4096)

	s.maxClientsCount = maxClientsAllowed
	s.conns.init(maxClientsAllowed)

	go s.loop()

	return nil
}

// Stop function
func (s *TCPServer) Stop() {
	s.stopCmdChan <- true
	<-s.exitLoopChan
	s.netListener = nil
}

func (s *TCPServer) loop() {
	defer s.netListener.Close()

	for {
		select {
		case <-s.stopCmdChan:
			s.exitLoopChan <- true
			return
		default:
			//s.netListener.SetDeadline(time.Now().Add(time.Millisecond))
			conn, err := s.netListener.AcceptTCP()
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else if err != nil {
				return
			}
			if s.conns.size() >= s.maxClientsCount {
				conn.Close()
				continue
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
func (s *TCPServer) Received(conn *Connection, packet *Packet) {
	s.callback.Received(conn, packet)
}

// Update function
func (s *TCPServer) Update() {
	for p := range s.packetsChan {
		s.callback.Received(p.conn, p.packet)
	}
}
