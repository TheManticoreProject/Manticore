package ip

import (
	"fmt"
	"strconv"
	"strings"
)

type IPv4 struct {
	A, B, C, D uint8
	MaskBits   uint8
}

// NewIPv4 creates a new IPv4 instance.
//
// Parameters:
//
//	a, b, c, d: The four octets of the IPv4 address.
//	maskBits: The number of bits in the subnet mask.
//
// Returns:
//
//	*IPv4: A new IPv4 instance.
func NewIPv4(a, b, c, d, maskBits uint8) *IPv4 {
	return &IPv4{
		A:        a,
		B:        b,
		C:        c,
		D:        d,
		MaskBits: maskBits,
	}
}

// NewIPv4FromString creates a new IPv4 instance from a string.
//
// Parameters:
//
//	s: The string representation of the IPv4 address.
//
// Returns:
//
//	*IPv4: A new IPv4 instance.
func NewIPv4FromString(s string) *IPv4 {
	parts := strings.Split(s, "/")

	if len(parts) == 2 {
		maskBits, err := strconv.ParseUint(parts[1], 10, 8)
		if err != nil {
			return nil
		}

		a, err := strconv.ParseUint(parts[0], 10, 8)
		if err != nil {
			return nil
		}

		b, err := strconv.ParseUint(parts[1], 10, 8)
		if err != nil {
			return nil
		}

		c, err := strconv.ParseUint(parts[2], 10, 8)
		if err != nil {
			return nil
		}

		d, err := strconv.ParseUint(parts[3], 10, 8)
		if err != nil {
			return nil
		}

		return NewIPv4(uint8(a), uint8(b), uint8(c), uint8(d), uint8(maskBits))
	}

	return nil
}

// ToUInt32 converts the IPv4 address to a 32-bit unsigned integer.
//
// Returns:
//
//	uint32: The 32-bit unsigned integer representation of the IPv4 address.
func (i *IPv4) ToUInt32() uint32 {
	return uint32(i.A)<<24 | uint32(i.B)<<16 | uint32(i.C)<<8 | uint32(i.D)
}

// ComputeMask calculates the network address by applying the subnet mask to the IPv4 address.
// It returns a new IPv4 instance representing the network address.
//
// The function first converts the IPv4 address to a 32-bit unsigned integer. Then, it creates
// a subnet mask by shifting the bits of 0xFFFFFFFF to the left by (32 - MaskBits) positions.
// The subnet mask is then applied to the 32-bit representation of the IPv4 address using a bitwise AND operation.
//
// Finally, the resulting network address is split back into its four octets (A, B, C, D) and
// a new IPv4 instance is created and returned with the same MaskBits value.
//
// Example:
//
//	ip := NewIPv4(192, 168, 1, 17, 24)
//	networkAddress := ip.ComputeMask()
//	fmt.Println(networkAddress.String()) // Output: 192.168.1.0/24
//
// Returns:
//
//	*IPv4: A new IPv4 instance representing the network address.
func (i *IPv4) ComputeMask() *IPv4 {
	n := i.ToUInt32()
	mask := uint32(0xFFFFFFFF) << (32 - i.MaskBits)
	masked := n & mask

	a := uint8((masked >> 24) & 0xFF)
	b := uint8((masked >> 16) & 0xFF)
	c := uint8((masked >> 8) & 0xFF)
	d := uint8(masked & 0xFF)

	return NewIPv4(a, b, c, d, i.MaskBits)
}

// String returns the string representation of the IPv4 address.
//
// Returns:
//
//	string: The string representation of the IPv4 address.
func (i *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d/%d", i.A, i.B, i.C, i.D, i.MaskBits)
}

// CIDRAddress returns the CIDR address of the IPv4 address.
//
// Returns:
//
//	string: The CIDR address of the IPv4 address.
func (i *IPv4) CIDRAddress() string {
	return fmt.Sprintf("%d.%d.%d.%d/%d", i.A, i.B, i.C, i.D, i.MaskBits)
}

// CIDRMask returns the CIDR mask of the IPv4 address.
//
// Returns:
//
//	string: The CIDR mask of the IPv4 address.
func (i *IPv4) CIDRMask() string {
	m := i.ComputeMask()
	return fmt.Sprintf("%d.%d.%d.%d/%d", m.A, m.B, m.C, m.D, m.MaskBits)
}

// IsInSubnet checks if the IPv4 address is within the subnet of another IPv4 address.
//
// Parameters:
//
//	subnet: The subnet to check against.
//
// Returns:
//
//	bool: True if the IPv4 address is within the subnet, false otherwise.
func (i *IPv4) IsInSubnet(subnet *IPv4) bool {
	return i.ToUInt32()&subnet.ToUInt32() == subnet.ToUInt32()
}

// IsInRange checks if the IPv4 address is within the range of two other IPv4 addresses.
//
// Parameters:
//
//	start: The start of the range.
//	end: The end of the range.
//
// Returns:
//
//	bool: True if the IPv4 address is within the range, false otherwise.
func (i *IPv4) IsInRange(start, end *IPv4) bool {
	return i.ToUInt32() >= start.ToUInt32() && i.ToUInt32() <= end.ToUInt32()
}
