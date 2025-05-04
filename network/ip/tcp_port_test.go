package ip_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ip"
)

func TestNewTCPPortRange(t *testing.T) {
	tests := []struct {
		start, end uint16
		expected   *ip.TCPPortRange
	}{
		{80, 8080, &ip.TCPPortRange{Start: 80, End: 8080}},
		{0, 65535, &ip.TCPPortRange{Start: 0, End: 65535}},
		{1024, 2048, &ip.TCPPortRange{Start: 1024, End: 2048}},
	}

	for _, tt := range tests {
		result := ip.NewTCPPortRange(tt.start, tt.end)
		if result.Start != tt.expected.Start || result.End != tt.expected.End {
			t.Errorf("NewTCPPortRange(%d, %d) = %v; expected %v", tt.start, tt.end, result, tt.expected)
		}
	}
}

func TestNewTCPPortRangeFromString(t *testing.T) {
	tests := []struct {
		s        string
		expected *ip.TCPPortRange
		hasError bool
	}{
		{"80-8080", &ip.TCPPortRange{Start: 80, End: 8080}, false},
		{"0-65535", &ip.TCPPortRange{Start: 0, End: 65535}, false},
		{"1024-2048", &ip.TCPPortRange{Start: 1024, End: 2048}, false},
		{"invalid", nil, true},
		{"80-70000", nil, true},
	}

	for _, tt := range tests {
		result, err := ip.NewTCPPortRangeFromString(tt.s)
		if (err != nil) != tt.hasError {
			t.Errorf("NewTCPPortRangeFromString(%s) error = %v; expected error = %v", tt.s, err, tt.hasError)
		}
		if err == nil && (result.Start != tt.expected.Start || result.End != tt.expected.End) {
			t.Errorf("NewTCPPortRangeFromString(%s) = %v; expected %v", tt.s, result, tt.expected)
		}
	}
}

func TestTCPPortRangeString(t *testing.T) {
	tests := []struct {
		portRange *ip.TCPPortRange
		expected  string
	}{
		{&ip.TCPPortRange{Start: 80, End: 8080}, "80-8080"},
		{&ip.TCPPortRange{Start: 0, End: 65535}, "0-65535"},
		{&ip.TCPPortRange{Start: 1024, End: 2048}, "1024-2048"},
	}

	for _, tt := range tests {
		result := tt.portRange.String()
		if result != tt.expected {
			t.Errorf("TCPPortRange.String() = %v; expected %v", result, tt.expected)
		}
	}
}
