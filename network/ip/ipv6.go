package ip

import (
	"fmt"
	"strconv"
	"strings"
)

type IPv6 struct {
	A, B, C, D, E, F, G, H uint16
}

// NewIPv6 creates a new IPv6 instance.
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
	}
}

// NewIPv6FromString creates a new IPv6 instance from a string.
//
// Parameters:
//
//	s: The string representation of the IPv6 address.
//
// Returns:
//
//	*IPv6: A new IPv6 instance.
func NewIPv6FromString(s string) *IPv6 {
	parts := strings.Split(s, ":")

	if len(parts) == 8 {
		a, err := strconv.ParseUint(parts[0], 16, 16)
		if err != nil {
			return nil
		}

		b, err := strconv.ParseUint(parts[1], 16, 16)
		if err != nil {
			return nil
		}

		c, err := strconv.ParseUint(parts[2], 16, 16)
		if err != nil {
			return nil
		}

		d, err := strconv.ParseUint(parts[3], 16, 16)
		if err != nil {
			return nil
		}

		e, err := strconv.ParseUint(parts[4], 16, 16)
		if err != nil {
			return nil
		}

		f, err := strconv.ParseUint(parts[5], 16, 16)
		if err != nil {
			return nil
		}

		g, err := strconv.ParseUint(parts[6], 16, 16)
		if err != nil {
			return nil
		}

		h, err := strconv.ParseUint(parts[7], 16, 16)
		if err != nil {
			return nil
		}

		return NewIPv6(uint16(a), uint16(b), uint16(c), uint16(d), uint16(e), uint16(f), uint16(g), uint16(h))
	}

	return nil
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
//
// Parameters:
//
//	subnet: The subnet to check against.
//
// Returns:
//
//	bool: True if the IPv6 address is within the subnet, false otherwise.
func (i *IPv6) IsInSubnet(subnet *IPv6) bool {
	return i.ToUInt128() == subnet.ToUInt128()
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
