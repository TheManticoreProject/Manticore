package ip_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ip"
)

func TestIPv4RangeContains(t *testing.T) {
	start := ip.NewIPv4(192, 168, 1, 0, 24)
	end := ip.NewIPv4(192, 168, 1, 255, 24)
	ipRange := &ip.IPv4Range{Start: start, End: end}

	tests := []struct {
		ip       *ip.IPv4
		expected bool
	}{
		{ip.NewIPv4(192, 168, 1, 1, 24), true},
		{ip.NewIPv4(192, 168, 1, 255, 24), true},
		{ip.NewIPv4(192, 168, 2, 0, 24), false},
		{ip.NewIPv4(192, 168, 0, 255, 24), false},
	}

	for _, test := range tests {
		result := ipRange.Contains(test.ip)
		if result != test.expected {
			t.Errorf("IPv4Range.Contains(%s) = %v; expected %v", test.ip.String(), result, test.expected)
		}
	}
}

func TestIPv6RangeContains(t *testing.T) {
	start := ip.NewIPv6(0x2001, 0xdb8, 0, 0, 0, 0, 0, 0)
	end := ip.NewIPv6(0x2001, 0xdb8, 0, 0, 0, 0, 0, 0xffff)
	ipRange := &ip.IPv6Range{Start: start, End: end}

	tests := []struct {
		ip       *ip.IPv6
		expected bool
	}{
		{ip.NewIPv6(0x2001, 0xdb8, 0, 0, 0, 0, 0, 1), true},
		{ip.NewIPv6(0x2001, 0xdb8, 0, 0, 0, 0, 0, 0xffff), true},
		{ip.NewIPv6(0x2001, 0xdb8, 0, 0, 0, 0, 1, 0), false},
		{ip.NewIPv6(0x2001, 0xdb7, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff, 0xffff), false},
	}

	for _, test := range tests {
		result := ipRange.Contains(test.ip)
		if result != test.expected {
			t.Errorf("IPv6Range.Contains(%s) = %v; expected %v", test.ip.String(), result, test.expected)
		}
	}
}
