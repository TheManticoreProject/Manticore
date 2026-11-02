package ip

import (
	"fmt"
	"iter"
)

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

// All iterates over the addresses of the range in ascending order, from Start to
// End inclusive, as /32 addresses. It yields nothing when End is before Start.
//
// Example:
//
//	for addr := range ip.NewIPv4FromString("10.0.0.0/30").CIDRRange().All() {
//	    fmt.Println(addr) // 10.0.0.0/32 … 10.0.0.3/32
//	}
func (r *IPv4Range) All() iter.Seq[*IPv4] {
	return func(yield func(*IPv4) bool) {
		start, end := r.Start.ToUInt32(), r.End.ToUInt32()
		if end < start {
			return
		}
		for v := start; ; v++ {
			if !yield(ipv4FromUint32(v)) || v == end {
				return
			}
		}
	}
}

// All iterates over the addresses of the range in ascending order, from Start to
// End inclusive, as /128 addresses. It yields nothing when End is before Start.
// IPv6 ranges can be astronomically large, so callers should stop early.
func (r *IPv6Range) All() iter.Seq[*IPv6] {
	return func(yield func(*IPv6) bool) {
		v, end := r.Start.ToUInt128(), r.End.ToUInt128()
		if v[0] > end[0] || (v[0] == end[0] && v[1] > end[1]) {
			return
		}
		for {
			if !yield(ipv6FromUint128(v, 128)) || v == end {
				return
			}
			v[1]++
			if v[1] == 0 {
				v[0]++
			}
		}
	}
}
