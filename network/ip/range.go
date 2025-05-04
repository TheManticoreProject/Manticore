package ip

import "fmt"

type IPv4Range struct {
	Start *IPv4
	End   *IPv4
}

// Contains checks if the given IPv4 address is within the range.
//
// Parameters:
//
//	ip: The IPv4 address to check.
//
// Returns:
//
//	bool: True if the IPv4 address is within the range, false otherwise.
func (r *IPv4Range) Contains(ip *IPv4) bool {
	return ip.IsInRange(r.Start, r.End)
}

// String returns the string representation of the IPv4 range.
//
// Returns:
//
//	string: The string representation of the IPv4 range.
func (r *IPv4Range) String() string {
	return fmt.Sprintf("%s - %s", r.Start.String(), r.End.String())
}

// IPv6Range represents a range of IPv6 addresses.
type IPv6Range struct {
	Start *IPv6
	End   *IPv6
}

// Contains checks if the given IPv6 address is within the range.
//
// Parameters:
//
//	ip: The IPv6 address to check.
//
// Returns:
//
//	bool: True if the IPv6 address is within the range, false otherwise.
func (r *IPv6Range) Contains(ip *IPv6) bool {
	return ip.IsInRange(r.Start, r.End)
}

// String returns the string representation of the IPv6 range.
//
// Returns:
//
//	string: The string representation of the IPv6 range.
func (r *IPv6Range) String() string {
	return fmt.Sprintf("%s - %s", r.Start.String(), r.End.String())
}
