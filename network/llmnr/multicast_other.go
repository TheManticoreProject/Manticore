//go:build !unix

package llmnr

import "net"

// setOutgoingMulticastInterface is a no-op on platforms where selecting the
// outgoing multicast interface via a raw socket option is not implemented. The
// link-local IPv6 destination still carries a zone (see multicastDestination),
// which the kernel uses to scope the send, so interface selection degrades
// gracefully rather than failing to build off the unix platforms.
func setOutgoingMulticastInterface(conn *net.UDPConn, family string, iface *net.Interface) error {
	return nil
}
