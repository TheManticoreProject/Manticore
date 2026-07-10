//go:build unix

package llmnr

import (
	"fmt"
	"net"
	"syscall"
)

// setOutgoingMulticastInterface pins iface as the interface out of which the
// client's multicast queries are sent. On a multi-homed host the kernel's
// default multicast egress interface is frequently not the one carrying the
// link's LLMNR traffic, so binding the outgoing interface explicitly is what
// makes interface selection effective.
//
// For IPv6 the option is IPV6_MULTICAST_IF (an interface index); for IPv4 it is
// IP_MULTICAST_IF, set via an ip_mreqn carrying the interface index so the
// interface can be named by index rather than by address. family is the network
// string the socket was created with ("udp4" or "udp6").
func setOutgoingMulticastInterface(conn *net.UDPConn, family string, iface *net.Interface) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	var sockErr error
	ctrlErr := raw.Control(func(fd uintptr) {
		switch family {
		case "udp6":
			sockErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_MULTICAST_IF, iface.Index)
		case "udp4":
			sockErr = syscall.SetsockoptIPMreqn(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_IF, &syscall.IPMreqn{Ifindex: int32(iface.Index)})
		default:
			sockErr = fmt.Errorf("unsupported address family %q", family)
		}
	})
	if ctrlErr != nil {
		return ctrlErr
	}
	return sockErr
}
