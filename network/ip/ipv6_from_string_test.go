package ip_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ip"
)

func TestNewIPv6FromString(t *testing.T) {
	valid := []struct {
		input string
		want  ip.IPv6
	}{
		{"1:2:3:4:5:6:7:8", ip.IPv6{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6, G: 7, H: 8}},
		{"fe80:0:0:0:0:0:0:1", ip.IPv6{A: 0xfe80, H: 1}},
		{"2001:0db8:0000:0000:0000:0000:0000:0001", ip.IPv6{A: 0x2001, B: 0xdb8, H: 1}},
		{"::", ip.IPv6{}},
		{"::1", ip.IPv6{H: 1}},
		{"fe80::1", ip.IPv6{A: 0xfe80, H: 1}},
		{"2001:db8::", ip.IPv6{A: 0x2001, B: 0xdb8}},
		{"2001:db8::8:800:200c:417a", ip.IPv6{A: 0x2001, B: 0xdb8, E: 8, F: 0x800, G: 0x200c, H: 0x417a}},
		{"FF01::101", ip.IPv6{A: 0xff01, H: 0x101}},
		{"::ffff:1.2.3.4", ip.IPv6{F: 0xffff, G: 0x0102, H: 0x0304}},
	}
	for _, tt := range valid {
		got := ip.NewIPv6FromString(tt.input)
		if got == nil {
			t.Errorf("NewIPv6FromString(%q) = nil, want %v", tt.input, tt.want.String())
			continue
		}
		if *got != tt.want {
			t.Errorf("NewIPv6FromString(%q) = %v, want %v", tt.input, got.String(), tt.want.String())
		}
		if again := ip.NewIPv6FromString(got.String()); again == nil || *again != *got {
			t.Errorf("NewIPv6FromString(%q) does not round-trip through String()", tt.input)
		}
	}

	invalid := []string{
		"",
		":",
		":::",
		"1::2::3",
		"1:2:3:4:5:6:7",
		"1:2:3:4:5:6:7:8:9",
		"1:2:3:4:5:6:7:10000",
		"00001:2:3:4:5:6:7:8",
		"fe80:0:0:0:0:0:0:1%eth0",
		"g::1",
		"1.2.3.4",
		"fe80::1%eth0",
		"::1/128",
	}
	for _, input := range invalid {
		if got := ip.NewIPv6FromString(input); got != nil {
			t.Errorf("NewIPv6FromString(%q) = %v, want nil", input, got.String())
		}
	}
}
