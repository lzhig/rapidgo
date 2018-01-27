package net

import (
	"net"
)

// TCPServer struct
type TCPServer struct {
	stopCmdChan  chan struct{}
	exitLoopChan chan struct{}

	maxClientsCount uint32
	conns           connections

	eventChan chan *Event
}

// Start function
func (s *TCPServer) Start(address string, maxClientsAllowed uint32) (<-chan *Event, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	netListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	s.stopCmdChan = make(chan struct{}, 1)
	s.exitLoopChan = make(chan struct{}, 1)
	s.eventChan = make(chan *Event, 1024)

	s.maxClientsCount = maxClientsAllowed
	s.conns.init(maxClientsAllowed)

	go s.loop(netListener)

	return s.eventChan, nil
}

// Stop function
func (s *TCPServer) Stop() {
	s.stopCmdChan <- struct{}{}
	<-s.exitLoopChan
}

func (s *TCPServer) loop(netListener *net.TCPListener) {
	defer netListener.Close()

	for {
		select {
		case <-s.stopCmdChan:
			s.exitLoopChan <- struct{}{}
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
			newConn.init()

			s.conns.add(newConn)
			newConn.packetHandler = config.PacketHandlerFactory(conn)
			s.eventChan <- &Event{Type: EventConnected, Conn: newConn}

			go newConn.loop(s.eventChan)
		}
	}
}
