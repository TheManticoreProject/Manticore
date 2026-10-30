package ip_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ip"
)

func TestNewIPv4FromString(t *testing.T) {
	valid := []struct {
		input string
		want  ip.IPv4
	}{
		{"192.168.1.10/24", ip.IPv4{A: 192, B: 168, C: 1, D: 10, MaskBits: 24}},
		{"0.0.0.0/0", ip.IPv4{MaskBits: 0}},
		{"255.255.255.255/32", ip.IPv4{A: 255, B: 255, C: 255, D: 255, MaskBits: 32}},
	}
	for _, tt := range valid {
		got := ip.NewIPv4FromString(tt.input)
		if got == nil {
			t.Errorf("NewIPv4FromString(%q) = nil, want %v", tt.input, tt.want.String())
			continue
		}
		if *got != tt.want {
			t.Errorf("NewIPv4FromString(%q) = %v, want %v", tt.input, got.String(), tt.want.String())
		}
		if again := ip.NewIPv4FromString(got.String()); again == nil || *again != *got {
			t.Errorf("NewIPv4FromString(%q) does not round-trip through String()", tt.input)
		}
	}

	invalid := []string{
		"",
		"1/2",
		"1.2.3/24",
		"1.2.3.4.5/24",
		"1.2.3.256/24",
		"1.2.3.4/33",
		"1.2.3.4/",
		"a.b.c.d/24",
		"1.2.3.4",
	}
	for _, input := range invalid {
		if got := ip.NewIPv4FromString(input); got != nil {
			t.Errorf("NewIPv4FromString(%q) = %v, want nil", input, got.String())
		}
	}
}
