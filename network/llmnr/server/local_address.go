package server

import (
	"fmt"
	"net"
)

// InterfaceAddresses returns the first usable IPv4 and IPv6 unicast addresses
// configured on the named interface, which a poisoner uses to auto-detect the
// address it should hand out so a coerced victim connects back to this host.
//
// When ifaceName is empty the addresses are taken from the first interface that
// is up, is not loopback, and carries a global-unicast address of the
// corresponding family. Either return value may be nil if the interface owns no
// address of that family; an error is returned only when no suitable interface
// or address could be found at all.
func InterfaceAddresses(ifaceName string) (ipv4 net.IP, ipv6 net.IP, err error) {
	if ifaceName != "" {
		iface, ierr := net.InterfaceByName(ifaceName)
		if ierr != nil {
			return nil, nil, fmt.Errorf("interface %q: %w", ifaceName, ierr)
		}
		ipv4, ipv6 = interfaceUnicastAddrs(iface)
		if ipv4 == nil && ipv6 == nil {
			return nil, nil, fmt.Errorf("interface %q has no usable unicast address", ifaceName)
		}
		return ipv4, ipv6, nil
	}

	ifaces, ierr := net.Interfaces()
	if ierr != nil {
		return nil, nil, fmt.Errorf("enumerating interfaces: %w", ierr)
	}
	for i := range ifaces {
		iface := ifaces[i]
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		v4, v6 := interfaceUnicastAddrs(&iface)
		if v4 != nil || v6 != nil {
			return v4, v6, nil
		}
	}
	return nil, nil, fmt.Errorf("no non-loopback interface with a usable unicast address was found")
}

// interfaceUnicastAddrs returns the first global-unicast IPv4 and IPv6 addresses
// configured on iface, or nil for a family the interface does not carry.
func interfaceUnicastAddrs(iface *net.Interface) (ipv4 net.IP, ipv6 net.IP) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, nil
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || !ipNet.IP.IsGlobalUnicast() {
			continue
		}
		if v4 := ipNet.IP.To4(); v4 != nil {
			if ipv4 == nil {
				ipv4 = v4
			}
			continue
		}
		if ipv6 == nil {
			ipv6 = ipNet.IP.To16()
		}
	}
	return ipv4, ipv6
}
