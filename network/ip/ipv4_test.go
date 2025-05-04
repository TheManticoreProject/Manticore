package ip_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ip"
)

func TestIPv4CIDRAddressString(t *testing.T) {
	tests := []struct {
		a, b, c, d uint8
		maskBits   uint8
		expected   string
	}{
		{192, 168, 1, 17, 24, "192.168.1.17/24"},
		{192, 168, 1, 28, 24, "192.168.1.28/24"},
		{192, 168, 1, 39, 24, "192.168.1.39/24"},
	}

	for _, tt := range tests {
		ipv4 := ip.NewIPv4(tt.a, tt.b, tt.c, tt.d, tt.maskBits)
		if result := ipv4.CIDRAddress(); result != tt.expected {
			t.Errorf("IPv4.CIDRAddress() = %v, want %v", result, tt.expected)
		}
	}
}

func TestIPv4CIDRMaskString(t *testing.T) {
	tests := []struct {
		a, b, c, d uint8
		maskBits   uint8
		expected   string
	}{
		{192, 168, 1, 17, 24, "192.168.1.0/24"},
		{192, 168, 1, 28, 24, "192.168.1.0/24"},
		{192, 168, 1, 39, 24, "192.168.1.0/24"},
	}

	for _, tt := range tests {
		ipv4 := ip.NewIPv4(tt.a, tt.b, tt.c, tt.d, tt.maskBits)
		if result := ipv4.CIDRMask(); result != tt.expected {
			t.Errorf("IPv4.CIDRMask() = %v, want %v", result, tt.expected)
		}
	}
}

func TestIPv4ToUInt32(t *testing.T) {
	tests := []struct {
		a, b, c, d uint8
		expected   uint32
	}{
		{192, 168, 1, 1, 3232235777},
		{10, 0, 0, 1, 167772161},
		{172, 16, 0, 1, 2886729729},
	}

	for _, tt := range tests {
		ipv4 := ip.NewIPv4(tt.a, tt.b, tt.c, tt.d, 24)
		if result := ipv4.ToUInt32(); result != tt.expected {
			t.Errorf("IPv4.ToUInt32() = %v, want %v", result, tt.expected)
		}
	}
}

func TestIPv4ComputeMask(t *testing.T) {
	tests := []struct {
		a, b, c, d uint8
		maskBits   uint8
		expected   string
	}{
		{192, 168, 1, 17, 24, "192.168.1.0/24"},
		{10, 0, 0, 1, 8, "10.0.0.0/8"},
		{172, 16, 0, 1, 12, "172.16.0.0/12"},
	}

	for _, tt := range tests {
		ipv4 := ip.NewIPv4(tt.a, tt.b, tt.c, tt.d, tt.maskBits)
		networkAddress := ipv4.ComputeMask()
		if result := networkAddress.String(); result != tt.expected {
			t.Errorf("IPv4.ComputeMask() = %v, want %v", result, tt.expected)
		}
	}
}

func TestIPv4String(t *testing.T) {
	tests := []struct {
		a, b, c, d uint8
		maskBits   uint8
		expected   string
	}{
		{192, 168, 1, 17, 24, "192.168.1.17/24"},
		{10, 0, 0, 1, 8, "10.0.0.1/8"},
		{172, 16, 0, 1, 12, "172.16.0.1/12"},
	}

	for _, tt := range tests {
		ipv4 := ip.NewIPv4(tt.a, tt.b, tt.c, tt.d, tt.maskBits)
		if result := ipv4.String(); result != tt.expected {
			t.Errorf("IPv4.String() = %v, want %v", result, tt.expected)
		}
	}
}

func TestIPv4IsInSubnet(t *testing.T) {
	tests := []struct {
		ipA, ipB *ip.IPv4
		expected bool
	}{
		{ip.NewIPv4(192, 168, 1, 17, 24), ip.NewIPv4(192, 168, 1, 0, 24), true},
		{ip.NewIPv4(10, 0, 0, 1, 8), ip.NewIPv4(10, 0, 0, 0, 8), true},
		{ip.NewIPv4(172, 16, 0, 1, 12), ip.NewIPv4(172, 16, 0, 0, 12), true},
		{ip.NewIPv4(192, 168, 2, 1, 24), ip.NewIPv4(192, 168, 1, 0, 24), false},
	}

	for _, tt := range tests {
		if result := tt.ipA.IsInSubnet(tt.ipB); result != tt.expected {
			t.Errorf("IPv4.IsInSubnet() = %v, want %v", result, tt.expected)
		}
	}
}

func TestIPv4IsInRange(t *testing.T) {
	tests := []struct {
		ipA, start, end *ip.IPv4
		expected        bool
	}{
		{ip.NewIPv4(192, 168, 1, 17, 24), ip.NewIPv4(192, 168, 1, 0, 24), ip.NewIPv4(192, 168, 1, 255, 24), true},
		{ip.NewIPv4(10, 0, 0, 1, 8), ip.NewIPv4(10, 0, 0, 0, 8), ip.NewIPv4(10, 255, 255, 255, 8), true},
		{ip.NewIPv4(172, 16, 0, 1, 12), ip.NewIPv4(172, 16, 0, 0, 12), ip.NewIPv4(172, 31, 255, 255, 12), true},
		{ip.NewIPv4(192, 168, 2, 1, 24), ip.NewIPv4(192, 168, 1, 0, 24), ip.NewIPv4(192, 168, 1, 255, 24), false},
	}

	for _, tt := range tests {
		if result := tt.ipA.IsInRange(tt.start, tt.end); result != tt.expected {
			t.Errorf("IPv4.IsInRange() = %v, want %v", result, tt.expected)
		}
	}
}
