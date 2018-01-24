package network

import "net"

// TCPConnection type
type TCPConnection struct {
	server *TCPServer
	rwc    net.Conn
}

func (obj *TCPConnection) serve() {

}
