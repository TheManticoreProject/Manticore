package message

import "github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"

// EndServerFrame is the one-way request 0x7006 that ends the Sign-on server
// job. The server sends no reply.
func EndServerFrame() hostserver.Frame {
	return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: RequestEndServer}}
}
