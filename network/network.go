package network

// CreateTCPServer function
func CreateTCPServer() *TCPServer {
	config := &TCPServerConfig{
		MaxConnections: 100,
		KeepAlive:      true,
	}
	return CreateTCPServerWithConfig(config)
}

// CreateTCPServerWithConfig function
func CreateTCPServerWithConfig(config *TCPServerConfig) *TCPServer {
	s := new(TCPServer)
	s.config = config
	return s
}
