package structures

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

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
	// FloorProtoNamedPipe identifies the named-pipe floor of an ncacn_np tower; its RHS
	// is the NUL-terminated pipe name (e.g. "\PIPE\srvsvc").
	FloorProtoNamedPipe = 0x0F
	// FloorProtoNetBIOS identifies the NetBIOS host-name floor (the address floor of an
	// ncacn_np tower); its RHS is the NUL-terminated host name.
	FloorProtoNetBIOS = 0x11
	// FloorProtoHTTP identifies the RPC-over-HTTP floor of an ncacn_http tower; its RHS is
	// a 16-bit port in big-endian order.
	FloorProtoHTTP = 0x1F
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

// nameFloor builds a single-identifier floor whose RHS is a NUL-terminated name (the
// shared shape of the named-pipe and NetBIOS floors).
func nameFloor(proto byte, name string) Floor {
	rhs := make([]byte, 0, len(name)+1)
	rhs = append(rhs, name...)
	rhs = append(rhs, 0) // NUL terminator
	return Floor{LHS: []byte{proto}, RHS: rhs}
}

// NamedPipeFloor builds the named-pipe floor of an ncacn_np tower; the RHS is the
// NUL-terminated pipe name (e.g. "\PIPE\srvsvc").
func NamedPipeFloor(name string) Floor { return nameFloor(FloorProtoNamedPipe, name) }

// NetBIOSFloor builds the NetBIOS host-name floor; the RHS is the NUL-terminated host
// name.
func NetBIOSFloor(host string) Floor { return nameFloor(FloorProtoNetBIOS, host) }

// HTTPFloor builds the RPC-over-HTTP floor; the RHS is the 16-bit port in big-endian
// order.
func HTTPFloor(port uint16) Floor {
	rhs := make([]byte, 2)
	binary.BigEndian.PutUint16(rhs, port)
	return Floor{LHS: []byte{FloorProtoHTTP}, RHS: rhs}
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
// floor. It is the ncacn_ip_tcp fast path used by ept_map's Map; for any other transport
// use Binding.
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

// BindingKind identifies the transport (protocol sequence) a tower describes.
type BindingKind uint8

const (
	// BindingUnknown marks a tower with no recognized transport/endpoint floor.
	BindingUnknown BindingKind = iota
	// BindingTCP is ncacn_ip_tcp (RPC over TCP).
	BindingTCP
	// BindingUDP is ncadg_ip_udp (connectionless RPC over UDP).
	BindingUDP
	// BindingNamedPipe is ncacn_np (RPC over an SMB named pipe).
	BindingNamedPipe
	// BindingHTTP is ncacn_http (RPC over HTTP / RPC proxy).
	BindingHTTP
)

// Binding is a tower decoded into the components of a DCE string binding
// (protseq:network_address[endpoint]).
type Binding struct {
	Kind BindingKind
	// ProtSeq is the protocol sequence, e.g. "ncacn_ip_tcp" or "ncacn_np".
	ProtSeq string
	// NetworkAddress is the host: an IPv4 address (ncacn_ip_tcp/http, ncadg_ip_udp) or a
	// NetBIOS host name (ncacn_np). It may be empty if the tower carries no address floor.
	NetworkAddress string
	// Endpoint is the per-transport endpoint: a decimal port (TCP/UDP/HTTP) or a pipe name
	// (named pipe).
	Endpoint string
}

// String renders the canonical DCE string binding, protseq:network_address[endpoint]
// (e.g. "ncacn_ip_tcp:10.0.0.1[49664]", "ncacn_np:HOST[\\PIPE\\srvsvc]").
func (b Binding) String() string {
	return fmt.Sprintf("%s:%s[%s]", b.ProtSeq, b.NetworkAddress, b.Endpoint)
}

// Binding decodes the tower's transport and address floors into a Binding. It supports
// ncacn_ip_tcp, ncadg_ip_udp, ncacn_np, and ncacn_http. ok-style failure is reported as an
// error when no recognized transport (endpoint) floor is present.
func (t Tower) Binding() (Binding, error) {
	var b Binding
	for _, f := range t.Floors {
		switch f.Protocol() {
		case FloorProtoTCP:
			b.Kind, b.ProtSeq, b.Endpoint = BindingTCP, "ncacn_ip_tcp", portString(f.RHS)
		case FloorProtoUDP:
			b.Kind, b.ProtSeq, b.Endpoint = BindingUDP, "ncadg_ip_udp", portString(f.RHS)
		case FloorProtoHTTP:
			b.Kind, b.ProtSeq, b.Endpoint = BindingHTTP, "ncacn_http", portString(f.RHS)
		case FloorProtoNamedPipe:
			b.Kind, b.ProtSeq, b.Endpoint = BindingNamedPipe, "ncacn_np", trimName(f.RHS)
		case FloorProtoIP:
			if len(f.RHS) >= 4 {
				b.NetworkAddress = net.IPv4(f.RHS[0], f.RHS[1], f.RHS[2], f.RHS[3]).String()
			}
		case FloorProtoNetBIOS:
			b.NetworkAddress = trimName(f.RHS)
		}
	}
	if b.Kind == BindingUnknown {
		return b, fmt.Errorf("epm: tower has no recognized transport floor")
	}
	return b, nil
}

// portString decodes a big-endian 16-bit port RHS to its decimal string, or "" if the RHS
// is too short.
func portString(rhs []byte) string {
	if len(rhs) < 2 {
		return ""
	}
	return strconv.Itoa(int(binary.BigEndian.Uint16(rhs)))
}

// trimName decodes a NUL-terminated name RHS (pipe or host) to a Go string.
func trimName(rhs []byte) string {
	return strings.TrimRight(string(rhs), "\x00")
}
