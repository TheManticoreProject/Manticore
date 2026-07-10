//go:build unix

package nbdgm

import (
	"net"
	"syscall"
)

// enableBroadcast sets the SO_BROADCAST socket option on conn so that datagrams
// addressed to the IPv4 limited-broadcast address (255.255.255.255) are allowed
// out. A BROADCAST datagram is sent to that address, and the kernel rejects a
// broadcast send on a socket that has not opted in via SO_BROADCAST.
func enableBroadcast(conn *net.UDPConn) error {
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	var sockErr error
	ctrlErr := raw.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	})
	if ctrlErr != nil {
		return ctrlErr
	}
	return sockErr
}
