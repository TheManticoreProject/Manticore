package constants_test

import (
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/constants"
)

// TestListenPort asserts the LLMNR UDP port matches RFC 4795 (5355).
func TestListenPort(t *testing.T) {
	if constants.ListenPort != 5355 {
		t.Errorf("ListenPort = %d, want 5355 (RFC 4795)", constants.ListenPort)
	}
}

// TestMulticastAddresses asserts the IPv4 and IPv6 LLMNR multicast group
// addresses match the values registered in RFC 4795 §2 and parse cleanly.
func TestMulticastAddresses(t *testing.T) {
	if constants.IPv4MulticastAddr != "224.0.0.252" {
		t.Errorf("IPv4MulticastAddr = %q, want %q", constants.IPv4MulticastAddr, "224.0.0.252")
	}
	ip4 := net.ParseIP(constants.IPv4MulticastAddr)
	if ip4 == nil || ip4.To4() == nil {
		t.Errorf("IPv4MulticastAddr = %q does not parse as an IPv4 address", constants.IPv4MulticastAddr)
	}
	if !ip4.IsMulticast() {
		t.Errorf("IPv4MulticastAddr = %q is not a multicast address", constants.IPv4MulticastAddr)
	}

	if constants.IPv6MulticastAddr != "FF02::1:3" {
		t.Errorf("IPv6MulticastAddr = %q, want %q", constants.IPv6MulticastAddr, "FF02::1:3")
	}
	ip6 := net.ParseIP(constants.IPv6MulticastAddr)
	if ip6 == nil {
		t.Fatalf("IPv6MulticastAddr = %q does not parse", constants.IPv6MulticastAddr)
	}
	if !ip6.IsMulticast() {
		t.Errorf("IPv6MulticastAddr = %q is not a multicast address", constants.IPv6MulticastAddr)
	}
}

// TestLabelAndDomainLimits asserts the DNS-derived label and name length limits.
func TestLabelAndDomainLimits(t *testing.T) {
	if constants.MaxLabelLength != 63 {
		t.Errorf("MaxLabelLength = %d, want 63", constants.MaxLabelLength)
	}
	if constants.MaxDomainLength != 255 {
		t.Errorf("MaxDomainLength = %d, want 255", constants.MaxDomainLength)
	}
}

// TestWireConstants asserts the DNS wire-format helper constants.
func TestWireConstants(t *testing.T) {
	if constants.LabelPointer != 0xC0 {
		t.Errorf("LabelPointer = %#x, want 0xC0", constants.LabelPointer)
	}
	if constants.MaxPacketSize != 512 {
		t.Errorf("MaxPacketSize = %d, want 512", constants.MaxPacketSize)
	}
}

// TestTimingConstants asserts the protocol timing constants from RFC 4795 §7.
func TestTimingConstants(t *testing.T) {
	if constants.JitterInterval != 100*time.Millisecond {
		t.Errorf("JitterInterval = %v, want %v", constants.JitterInterval, 100*time.Millisecond)
	}
	if constants.LLMNRTimeout != 1*time.Second {
		t.Errorf("LLMNRTimeout = %v, want %v", constants.LLMNRTimeout, 1*time.Second)
	}
}
