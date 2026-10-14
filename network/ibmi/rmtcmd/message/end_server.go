package message

import "github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"

// EndServerFrame is the one-way request 0x1004 that ends the Remote Command
// server job. Correlation zero is reserved for startup and end exchanges.
func EndServerFrame() hostserver.Frame {
	return hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDRemoteCommand, RequestID: RequestEndServer}}
}
