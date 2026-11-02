package ip

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strings"
)

type IPv6 struct {
	A, B, C, D, E, F, G, H uint16
	// MaskBits is the prefix length (0-128). NewIPv6 and NewIPv6FromString set
	// 128 for a plain address, so IsInSubnet against it matches that address
	// only; a zero value is a /0 prefix that contains every address.
	MaskBits uint8
}

// NewIPv6 creates a new IPv6 instance with a /128 prefix.
//
// Parameters:
//
//	a, b, c, d, e, f, g, h: The eight segments of the IPv6 address.
//
// Returns:
//
//	*IPv6: A new IPv6 instance.
func NewIPv6(a, b, c, d, e, f, g, h uint16) *IPv6 {
	return &IPv6{
		A: a, B: b, C: c, D: d, E: e, F: f, G: g, H: h,
		MaskBits: 128,
	}
}

// ipv6FromUint128 builds an IPv6 from its 128-bit value and a prefix length.
func ipv6FromUint128(v [2]uint64, maskBits uint8) *IPv6 {
	return &IPv6{
		A: uint16(v[0] >> 48), B: uint16(v[0] >> 32), C: uint16(v[0] >> 16), D: uint16(v[0]),
		E: uint16(v[1] >> 48), F: uint16(v[1] >> 32), G: uint16(v[1] >> 16), H: uint16(v[1]),
		MaskBits: maskBits,
	}
}

// ipv6Mask returns the 128-bit netmask of a prefix length (clamped to 128).
func ipv6Mask(maskBits uint8) [2]uint64 {
	switch {
	case maskBits == 0:
		return [2]uint64{0, 0}
	case maskBits <= 64:
		return [2]uint64{^uint64(0) << (64 - maskBits), 0}
	case maskBits < 128:
		return [2]uint64{^uint64(0), ^uint64(0) << (128 - maskBits)}
	default:
		return [2]uint64{^uint64(0), ^uint64(0)}
	}
}

// NewIPv6FromString creates a new IPv6 instance from a string.
//
// Both the full form ("2001:db8:0:0:0:0:0:1") and the compressed "::"
// form of RFC 4291 section 2.2 ("2001:db8::1", "::1", "::") are accepted,
// as well as an embedded IPv4 suffix ("::ffff:1.2.3.4"), optionally followed
// by a prefix length ("2001:db8::/32", RFC 4291 section 2.3). A plain address
// gets a /128 prefix. Plain IPv4 addresses and addresses carrying a zone
// ("fe80::1%eth0") are rejected.
//
// Parameters:
//
//	s: The string representation of the IPv6 address.
//
// Returns:
//
//	*IPv6: A new IPv6 instance, or nil if s is not a valid IPv6 address.
func NewIPv6FromString(s string) *IPv6 {
	maskBits := 128
	if strings.Contains(s, "/") {
		prefix, err := netip.ParsePrefix(s)
		if err != nil {
			return nil
		}
		maskBits = prefix.Bits()
		s = prefix.Addr().String()
	}
	addr, err := netip.ParseAddr(s)
	if err != nil || !addr.Is6() || addr.Zone() != "" {
		return nil
	}

	b := addr.As16()
	ip := NewIPv6(
		binary.BigEndian.Uint16(b[0:2]),
		binary.BigEndian.Uint16(b[2:4]),
		binary.BigEndian.Uint16(b[4:6]),
		binary.BigEndian.Uint16(b[6:8]),
		binary.BigEndian.Uint16(b[8:10]),
		binary.BigEndian.Uint16(b[10:12]),
		binary.BigEndian.Uint16(b[12:14]),
		binary.BigEndian.Uint16(b[14:16]),
	)
	ip.MaskBits = uint8(maskBits)
	return ip
}

// ToUInt128 converts the IPv6 address to a 128-bit unsigned integer.
//
// Returns:
//
//	[2]uint64: The 128-bit unsigned integer representation of the IPv6 address.
func (i *IPv6) ToUInt128() [2]uint64 {
	high := uint64(i.A)<<48 | uint64(i.B)<<32 | uint64(i.C)<<16 | uint64(i.D)
	low := uint64(i.E)<<48 | uint64(i.F)<<32 | uint64(i.G)<<16 | uint64(i.H)
	return [2]uint64{high, low}
}

// String returns the string representation of the IPv6 address.
//
// Returns:
//
//	string: The string representation of the IPv6 address.
func (i *IPv6) String() string {
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x", i.A, i.B, i.C, i.D, i.E, i.F, i.G, i.H)
}

// IsInSubnet checks if the IPv6 address is within the subnet of another IPv6 address.
// The subnet's MaskBits selects how many leading bits must match; a subnet with a
// prefix longer than 128 contains nothing.
//
// Parameters:
//
//	subnet: The subnet to check against.
//
// Returns:
//
//	bool: True if the IPv6 address is within the subnet, false otherwise.
func (i *IPv6) IsInSubnet(subnet *IPv6) bool {
	if subnet.MaskBits > 128 {
		return false
	}
	mask := ipv6Mask(subnet.MaskBits)
	ip, sn := i.ToUInt128(), subnet.ToUInt128()
	return ip[0]&mask[0] == sn[0]&mask[0] && ip[1]&mask[1] == sn[1]&mask[1]
}

// ComputeMask returns the network address of the prefix: the address with every
// bit past MaskBits cleared, keeping MaskBits.
func (i *IPv6) ComputeMask() *IPv6 {
	mask := ipv6Mask(i.MaskBits)
	v := i.ToUInt128()
	return ipv6FromUint128([2]uint64{v[0] & mask[0], v[1] & mask[1]}, i.MaskBits)
}

// CIDRAddress returns the address with its prefix length, e.g. "2001:db8::1/64".
func (i *IPv6) CIDRAddress() string {
	return fmt.Sprintf("%s/%d", i.String(), i.MaskBits)
}

// CIDRRange returns the range of addresses the prefix covers, from its network
// address to the address with every host bit set. Both ends are /128 addresses.
func (i *IPv6) CIDRRange() *IPv6Range {
	mask := ipv6Mask(i.MaskBits)
	v := i.ToUInt128()
	start := [2]uint64{v[0] & mask[0], v[1] & mask[1]}
	end := [2]uint64{start[0] | ^mask[0], start[1] | ^mask[1]}
	return &IPv6Range{Start: ipv6FromUint128(start, 128), End: ipv6FromUint128(end, 128)}
}

// IsInRange checks if the IPv6 address is within the range of two other IPv6 addresses.
//
// Parameters:
//
//	start: The start of the range.
//	end: The end of the range.
//
// Returns:
//
//	bool: True if the IPv6 address is within the range, false otherwise.
func (i *IPv6) IsInRange(start, end *IPv6) bool {
	ip := i.ToUInt128()
	startIP := start.ToUInt128()
	endIP := end.ToUInt128()
	return (ip[0] > startIP[0] || (ip[0] == startIP[0] && ip[1] >= startIP[1])) &&
		(ip[0] < endIP[0] || (ip[0] == endIP[0] && ip[1] <= endIP[1]))
}
