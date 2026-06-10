package structures

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// A protocol tower is the floor-based description of an interface-over-transport
// binding, defined in [C706] Appendix L with the floor protocol identifiers in [C706]
// Appendix I. It is carried in ept_map as an octet string inside a twr_t (see Twr); the
// encoding here is that raw octet string, independent of NDR.
//
// References:
//   - [C706] Appendix L (protocol tower encoding):
//     https://pubs.opengroup.org/onlinepubs/9629399/apdxl.htm
//   - [C706] Appendix I (protocol identifiers):
//     https://pubs.opengroup.org/onlinepubs/9629399/apdxi.htm

// Floor protocol identifier bytes, the first octet of a floor's LHS ([C706] Appendix I).
const (
	// FloorProtoUUID marks a floor whose LHS is a UUID plus a 16-bit major version (the
	// interface-identifier and transfer-syntax floors). The RHS holds the 16-bit minor
	// version.
	FloorProtoUUID = 0x0D
	// FloorProtoNCACN identifies the connection-oriented RPC protocol (major v5).
	FloorProtoNCACN = 0x0B
	// FloorProtoNCADG identifies the connectionless RPC protocol (major v4).
	FloorProtoNCADG = 0x0A
	// FloorProtoTCP identifies the DOD TCP floor; its RHS is a 16-bit port in big-endian
	// order. This is the transport floor of an ncacn_ip_tcp tower.
	FloorProtoTCP = 0x07
	// FloorProtoUDP identifies the DOD UDP floor; its RHS is a 16-bit port in big-endian
	// order.
	FloorProtoUDP = 0x08
	// FloorProtoIP identifies the DOD IP floor; its RHS is a 4-octet IPv4 address in
	// big-endian order.
	FloorProtoIP = 0x09
)

// Floor is one floor of a protocol tower: a length-prefixed left-hand side (protocol
// identifier and associated data) and a length-prefixed right-hand side (addressing or
// version data).
type Floor struct {
	LHS []byte
	RHS []byte
}

// Protocol returns the floor's protocol identifier byte (the first LHS octet), or 0 if
// the LHS is empty.
func (f Floor) Protocol() byte {
	if len(f.LHS) == 0 {
		return 0
	}
	return f.LHS[0]
}

// Marshal serializes the floor: uint16 LHS length, LHS, uint16 RHS length, RHS (all
// lengths little-endian).
func (f Floor) Marshal() []byte {
	buf := make([]byte, 0, 4+len(f.LHS)+len(f.RHS))
	var l [2]byte
	binary.LittleEndian.PutUint16(l[:], uint16(len(f.LHS)))
	buf = append(buf, l[:]...)
	buf = append(buf, f.LHS...)
	binary.LittleEndian.PutUint16(l[:], uint16(len(f.RHS)))
	buf = append(buf, l[:]...)
	buf = append(buf, f.RHS...)
	return buf
}

// unmarshalFloor parses one floor from the front of data and returns it with the number
// of bytes consumed.
func unmarshalFloor(data []byte) (Floor, int, error) {
	if len(data) < 2 {
		return Floor{}, 0, fmt.Errorf("epm: floor truncated reading LHS length")
	}
	off := 0
	lhsLen := int(binary.LittleEndian.Uint16(data[off:]))
	off += 2
	if len(data) < off+lhsLen+2 {
		return Floor{}, 0, fmt.Errorf("epm: floor truncated reading LHS/RHS length")
	}
	lhs := append([]byte(nil), data[off:off+lhsLen]...)
	off += lhsLen
	rhsLen := int(binary.LittleEndian.Uint16(data[off:]))
	off += 2
	if len(data) < off+rhsLen {
		return Floor{}, 0, fmt.Errorf("epm: floor truncated reading RHS")
	}
	rhs := append([]byte(nil), data[off:off+rhsLen]...)
	off += rhsLen
	return Floor{LHS: lhs, RHS: rhs}, off, nil
}

// Tower is an ordered list of floors describing an interface-over-transport binding.
type Tower struct {
	Floors []Floor
}

// Marshal serializes the tower: a little-endian uint16 floor count followed by each
// floor.
func (t Tower) Marshal() []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(t.Floors)))
	for _, f := range t.Floors {
		buf = append(buf, f.Marshal()...)
	}
	return buf
}

// UnmarshalTower parses a tower octet string.
func UnmarshalTower(data []byte) (Tower, error) {
	if len(data) < 2 {
		return Tower{}, fmt.Errorf("epm: tower truncated reading floor count")
	}
	count := int(binary.LittleEndian.Uint16(data))
	off := 2
	t := Tower{Floors: make([]Floor, 0, count)}
	for i := 0; i < count; i++ {
		f, n, err := unmarshalFloor(data[off:])
		if err != nil {
			return Tower{}, fmt.Errorf("epm: floor %d: %w", i, err)
		}
		t.Floors = append(t.Floors, f)
		off += n
	}
	return t, nil
}

// uuidFloor builds a UUID-typed floor (interface identifier or transfer syntax): LHS =
// 0x0D, 16-byte UUID, 2-byte major version; RHS = 2-byte minor version.
func uuidFloor(id guid.GUID, major, minor uint16) Floor {
	lhs := make([]byte, 0, 1+16+2)
	lhs = append(lhs, FloorProtoUUID)
	lhs = append(lhs, id.ToBytes()...)
	var v [2]byte
	binary.LittleEndian.PutUint16(v[:], major)
	lhs = append(lhs, v[:]...)
	rhs := make([]byte, 2)
	binary.LittleEndian.PutUint16(rhs, minor)
	return Floor{LHS: lhs, RHS: rhs}
}

// InterfaceFloor builds the interface-identifier floor (floor 1).
func InterfaceFloor(iface guid.GUID, major, minor uint16) Floor {
	return uuidFloor(iface, major, minor)
}

// TransferSyntaxFloor builds the NDR transfer-syntax floor (floor 2).
func TransferSyntaxFloor() Floor {
	nd := syntax.NDRTransferSyntax()
	return uuidFloor(nd.UUID, nd.MajorVersion, nd.MinorVersion)
}

// protocolFloor builds a single-identifier RPC protocol floor (LHS = one identifier
// byte) whose RHS is a 2-byte little-endian minor version.
func protocolFloor(proto byte, minor uint16) Floor {
	rhs := make([]byte, 2)
	binary.LittleEndian.PutUint16(rhs, minor)
	return Floor{LHS: []byte{proto}, RHS: rhs}
}

// TCPFloor builds the DOD TCP floor; the RHS is the 16-bit port in big-endian order.
func TCPFloor(port uint16) Floor {
	rhs := make([]byte, 2)
	binary.BigEndian.PutUint16(rhs, port)
	return Floor{LHS: []byte{FloorProtoTCP}, RHS: rhs}
}

// IPFloor builds the DOD IP floor; the RHS is the 4-octet IPv4 address in big-endian
// (network) order.
func IPFloor(ip net.IP) Floor {
	rhs := make([]byte, 4)
	if v4 := ip.To4(); v4 != nil {
		copy(rhs, v4)
	}
	return Floor{LHS: []byte{FloorProtoIP}, RHS: rhs}
}

// BuildMapTowerTCP builds the 5-floor ncacn_ip_tcp tower used as the ept_map input: the
// interface and NDR transfer-syntax floors describe what is wanted, and the
// connection-oriented/TCP/IP floors with a zero port and address request that the
// endpoint mapper fill in the bound endpoint.
func BuildMapTowerTCP(iface guid.GUID, ifMajor, ifMinor uint16) Tower {
	return Tower{Floors: []Floor{
		InterfaceFloor(iface, ifMajor, ifMinor),
		TransferSyntaxFloor(),
		protocolFloor(FloorProtoNCACN, 0),
		TCPFloor(0),
		IPFloor(net.IPv4zero),
	}}
}

// Endpoint is a resolved transport endpoint extracted from a tower.
type Endpoint struct {
	IP   net.IP
	Port uint16
}

// String renders the endpoint as host:port.
func (e Endpoint) String() string { return fmt.Sprintf("%s:%d", e.IP, e.Port) }

// Endpoint extracts the TCP port and IPv4 address from the tower's transport floors. ok
// is false unless a TCP floor is present; the IP is left zero if the tower carries no IP
// floor.
func (t Tower) Endpoint() (ep Endpoint, ok bool) {
	var gotPort bool
	for _, f := range t.Floors {
		switch f.Protocol() {
		case FloorProtoTCP:
			if len(f.RHS) >= 2 {
				ep.Port = binary.BigEndian.Uint16(f.RHS)
				gotPort = true
			}
		case FloorProtoIP:
			if len(f.RHS) >= 4 {
				ep.IP = net.IPv4(f.RHS[0], f.RHS[1], f.RHS[2], f.RHS[3])
			}
		}
	}
	return ep, gotPort
}
