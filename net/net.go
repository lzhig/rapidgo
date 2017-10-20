package net

// ICallback interface
type ICallback interface {
	Disconnected(conn *Connection, err error)
	Connected(conn *Connection)
	Received(conn *Connection, packet Packet)
}
