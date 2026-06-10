package transport

import (
	"net"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
	"github.com/TheManticoreProject/Manticore/network/tcp"
)

type Transport interface {
	Connect(ipaddr net.IP, port int) error

	Close() error

	Send(data []byte) (int, error)

	Receive() ([]byte, error)

	IsConnected() bool

	// SetTimeout bounds Connect and each subsequent Receive: Connect fails if
	// the connection cannot be established within d, and Receive fails if a
	// frame does not arrive within d. A non-positive d removes the bound
	// (blocking I/O), which is the initial state of every transport.
	SetTimeout(d time.Duration)
}

func NewTransport(transportType string) Transport {
	switch strings.ToLower(transportType) {
	case "tcp":
		return tcp.NewTCPTransport()
	case "nbt":
		return nbt.NewNBTTransport()
	}
	return nil
}
