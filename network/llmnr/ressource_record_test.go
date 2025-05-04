package llmnr_test

import (
	"github.com/TheManticoreProject/Manticore/network/llmnr"

	"bytes"
	"testing"
)

func TestIPToRData(t *testing.T) {
	tests := []struct {
		ip       string
		expected []byte
	}{
		{"192.168.1.1", []byte{192, 168, 1, 1}},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", []byte{32, 1, 13, 184, 133, 163, 0, 0, 0, 0, 138, 46, 3, 112, 115, 52}},
		{"invalid_ip", nil},
	}

	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			result := llmnr.IPToRData(test.ip)
			if result == nil && test.expected == nil {
				return
			} else if !bytes.Equal(result, test.expected) {
				t.Errorf("IPToRData(%s) = %v; want %v", test.ip, result, test.expected)
			}
		})
	}
}

func TestIPv4ToRData(t *testing.T) {
	tests := []struct {
		ip       string
		expected []byte
	}{
		{"192.168.1.1", []byte{192, 168, 1, 1}},
		{"0.0.0.0", []byte{0, 0, 0, 0}},
		{"255.255.255.255", []byte{255, 255, 255, 255}},
		{"127.0.0.1", []byte{127, 0, 0, 1}},
		{"", nil},
		{"256.256.256.256", nil},
		{"invalid_ip", nil},
	}

	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			result := llmnr.IPv4ToRData(test.ip)
			if result == nil && test.expected == nil {
				return
			} else if !bytes.Equal(result, test.expected) {
				t.Errorf("IPv4ToRData(%s) = %v; want %v", test.ip, result, test.expected)
			}
		})
	}
}

func TestIPv6ToRData(t *testing.T) {
	tests := []struct {
		ip       string
		expected []byte
	}{
		{"::1", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"::", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"2001:db8::", []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"2001:db8::1", []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"2001:db8:0:0:0:0:2:1", []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}},
		{"2001:db8:0:0:0:0:0:1", []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"2001:db8::2:1", []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", []byte{0x20, 0x01, 0x0d, 0xb8, 0x85, 0xa3, 0x00, 0x00, 0x00, 0x00, 0x8a, 0x2e, 0x03, 0x70, 0x73, 0x34}},
		{"invalid_ip", nil},
	}

	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			result := llmnr.IPv6ToRData(test.ip)
			if result == nil && test.expected == nil {
				return
			} else if !bytes.Equal(result, test.expected) {
				t.Errorf("IPv6ToRData(%s) = %v; want %v", test.ip, result, test.expected)
			}
		})
	}
}
