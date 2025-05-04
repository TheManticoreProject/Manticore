package transport

import (
	"net"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
)

type Transport interface {
	Connect(ipaddr net.IP, port int) error

	Close() error

	Send(data []byte) (int, error)

	Receive() ([]byte, error)

	IsConnected() bool
}

func NewTransport(transportType string) Transport {
	switch strings.ToLower(transportType) {
	case "nbt":
		return nbt.NewNBTTransport()
	}
	return nil
}
