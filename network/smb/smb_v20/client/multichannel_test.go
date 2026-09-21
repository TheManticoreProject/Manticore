package client

import (
	"encoding/binary"
	"net"
	"testing"
)

func buildNetworkInterfaceEntry(ifIndex uint32, cap uint32, linkSpeed uint64, family uint16, ip net.IP, port uint16, next uint32) []byte {
	entry := make([]byte, 152)
	binary.LittleEndian.PutUint32(entry[0:4], next)
	binary.LittleEndian.PutUint32(entry[4:8], ifIndex)
	binary.LittleEndian.PutUint32(entry[8:12], cap)
	// entry[12:16] reserved
	binary.LittleEndian.PutUint64(entry[16:24], linkSpeed)

	sockaddr := entry[24:152]
	binary.LittleEndian.PutUint16(sockaddr[0:2], family)
	binary.BigEndian.PutUint16(sockaddr[2:4], port)

	switch family {
	case 2: // AF_INET
		copy(sockaddr[4:8], ip.To4())
	case 23: // AF_INET6
		copy(sockaddr[8:24], ip.To16())
	}
	return entry
}

func TestParseNetworkInterfacesSingleIPv4(t *testing.T) {
	entry := buildNetworkInterfaceEntry(1, INTERFACE_CAP_RSS, 10_000_000_000, 2, net.IPv4(10, 0, 0, 1), 445, 0)

	ifaces, err := parseNetworkInterfaces(entry)
	if err != nil {
		t.Fatalf("parseNetworkInterfaces: %v", err)
	}
	if len(ifaces) != 1 {
		t.Fatalf("got %d interfaces, want 1", len(ifaces))
	}
	ni := ifaces[0]
	if ni.IfIndex != 1 {
		t.Errorf("IfIndex = %d, want 1", ni.IfIndex)
	}
	if !ni.IsRSS() {
		t.Error("expected RSS capability")
	}
	if ni.IsRDMA() {
		t.Error("unexpected RDMA capability")
	}
	if ni.LinkSpeed != 10_000_000_000 {
		t.Errorf("LinkSpeed = %d, want 10000000000", ni.LinkSpeed)
	}
	if !ni.SockAddrInet.Equal(net.IPv4(10, 0, 0, 1)) {
		t.Errorf("SockAddrInet = %s, want 10.0.0.1", ni.SockAddrInet)
	}
	if ni.SockAddrPort != 445 {
		t.Errorf("SockAddrPort = %d, want 445", ni.SockAddrPort)
	}
}

func TestParseNetworkInterfacesSingleIPv6(t *testing.T) {
	ip6 := net.ParseIP("fe80::1")
	entry := buildNetworkInterfaceEntry(2, 0, 1_000_000_000, 23, ip6, 445, 0)

	ifaces, err := parseNetworkInterfaces(entry)
	if err != nil {
		t.Fatalf("parseNetworkInterfaces: %v", err)
	}
	if len(ifaces) != 1 {
		t.Fatalf("got %d interfaces, want 1", len(ifaces))
	}
	if !ifaces[0].SockAddrInet.Equal(ip6) {
		t.Errorf("SockAddrInet = %s, want %s", ifaces[0].SockAddrInet, ip6)
	}
	if ifaces[0].Family != 23 {
		t.Errorf("Family = %d, want 23", ifaces[0].Family)
	}
}

func TestParseNetworkInterfacesChained(t *testing.T) {
	e1 := buildNetworkInterfaceEntry(1, INTERFACE_CAP_RSS, 10_000_000_000, 2, net.IPv4(10, 0, 0, 1), 445, 152)
	e2 := buildNetworkInterfaceEntry(2, INTERFACE_CAP_RDMA, 40_000_000_000, 2, net.IPv4(10, 0, 0, 2), 445, 0)
	data := append(e1, e2...)

	ifaces, err := parseNetworkInterfaces(data)
	if err != nil {
		t.Fatalf("parseNetworkInterfaces: %v", err)
	}
	if len(ifaces) != 2 {
		t.Fatalf("got %d interfaces, want 2", len(ifaces))
	}
	if !ifaces[0].SockAddrInet.Equal(net.IPv4(10, 0, 0, 1)) {
		t.Errorf("iface[0] IP = %s, want 10.0.0.1", ifaces[0].SockAddrInet)
	}
	if !ifaces[1].SockAddrInet.Equal(net.IPv4(10, 0, 0, 2)) {
		t.Errorf("iface[1] IP = %s, want 10.0.0.2", ifaces[1].SockAddrInet)
	}
	if !ifaces[1].IsRDMA() {
		t.Error("iface[1] expected RDMA capability")
	}
}

func TestParseNetworkInterfacesTooShort(t *testing.T) {
	_, err := parseNetworkInterfaces(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestQueryNetworkInterfaceInfoRequiresTree(t *testing.T) {
	ft := &fakeTransport{connected: true}
	c := newTestClient(ft)

	_, err := c.QueryNetworkInterfaceInfo()
	if err == nil {
		t.Error("expected error when no tree connect is established")
	}
}
