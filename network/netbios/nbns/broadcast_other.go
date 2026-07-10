//go:build !unix

package nbns

import "net"

// enableBroadcast is a no-op on platforms where toggling SO_BROADCAST through a
// raw socket option is not wired up here. Unicast queries to an NBNS/WINS server
// are unaffected; only the limited-broadcast (B-node) send depends on the
// option, and it degrades gracefully rather than failing to build off unix.
func enableBroadcast(conn *net.UDPConn) error {
	return nil
}
