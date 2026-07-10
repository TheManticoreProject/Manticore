package resourcerecord_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

// buildRecord marshals a ResourceRecord and unmarshals it back through the
// message-relative path, returning the decoded record. The record is placed at
// a non-zero offset inside a synthetic buffer so that any accidental use of a
// sub-slice origin (rather than the message origin) for RDATA name decoding
// would surface as a wrong result.
func buildRecord(t *testing.T, name string, rtype llmnr_type.Type, rdata []byte) resourcerecord.ResourceRecord {
	t.Helper()

	rr := resourcerecord.ResourceRecord{
		Name:  domain_name.DomainName(name),
		Type:  rtype,
		Class: class.ClassIN,
		TTL:   30,
		RData: rdata,
	}
	wire, err := rr.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded resourcerecord.ResourceRecord
	n, err := decoded.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(wire) {
		t.Fatalf("consumed %d bytes, want %d", n, len(wire))
	}
	return decoded
}

func TestAsARoundTrip(t *testing.T) {
	rdata := resourcerecord.IPv4ToRData("10.7.0.10")
	rr := resourcerecord.ResourceRecord{Type: llmnr_type.TypeA, RData: rdata}

	ip, err := rr.AsA()
	if err != nil {
		t.Fatalf("AsA failed: %v", err)
	}
	if !ip.Equal(net.ParseIP("10.7.0.10")) {
		t.Errorf("AsA = %v, want 10.7.0.10", ip)
	}

	// Wrong type must be rejected.
	rr.Type = llmnr_type.TypeAAAA
	if _, err := rr.AsA(); err == nil {
		t.Errorf("AsA on non-A record should error")
	}

	// Wrong length must be rejected.
	rr = resourcerecord.ResourceRecord{Type: llmnr_type.TypeA, RData: []byte{1, 2, 3}}
	if _, err := rr.AsA(); err == nil {
		t.Errorf("AsA with 3-byte rdata should error")
	}
}

func TestAsAAAARoundTrip(t *testing.T) {
	rdata := resourcerecord.IPv6ToRData("2001:db8::1")
	rr := resourcerecord.ResourceRecord{Type: llmnr_type.TypeAAAA, RData: rdata}

	ip, err := rr.AsAAAA()
	if err != nil {
		t.Fatalf("AsAAAA failed: %v", err)
	}
	if !ip.Equal(net.ParseIP("2001:db8::1")) {
		t.Errorf("AsAAAA = %v, want 2001:db8::1", ip)
	}

	rr = resourcerecord.ResourceRecord{Type: llmnr_type.TypeAAAA, RData: []byte{1, 2, 3}}
	if _, err := rr.AsAAAA(); err == nil {
		t.Errorf("AsAAAA with short rdata should error")
	}
}

func TestAsPTRRoundTrip(t *testing.T) {
	rdata, err := resourcerecord.NameToRData("host.example.com")
	if err != nil {
		t.Fatalf("NameToRData failed: %v", err)
	}
	rr := buildRecord(t, "10.0.0.1.in-addr.arpa", llmnr_type.TypePTR, rdata)

	name, err := rr.AsPTR()
	if err != nil {
		t.Fatalf("AsPTR failed: %v", err)
	}
	if name != "host.example.com" {
		t.Errorf("AsPTR = %q, want %q", name, "host.example.com")
	}
}

func TestAsCNAMERoundTrip(t *testing.T) {
	rdata, err := resourcerecord.NameToRData("canonical.example.com")
	if err != nil {
		t.Fatalf("NameToRData failed: %v", err)
	}
	rr := buildRecord(t, "alias.example.com", llmnr_type.TypeCNAME, rdata)

	name, err := rr.AsCNAME()
	if err != nil {
		t.Fatalf("AsCNAME failed: %v", err)
	}
	if name != "canonical.example.com" {
		t.Errorf("AsCNAME = %q, want %q", name, "canonical.example.com")
	}
}

func TestAsNSRoundTrip(t *testing.T) {
	rdata, err := resourcerecord.NameToRData("ns1.example.com")
	if err != nil {
		t.Fatalf("NameToRData failed: %v", err)
	}
	rr := buildRecord(t, "example.com", llmnr_type.TypeNS, rdata)

	name, err := rr.AsNS()
	if err != nil {
		t.Fatalf("AsNS failed: %v", err)
	}
	if name != "ns1.example.com" {
		t.Errorf("AsNS = %q, want %q", name, "ns1.example.com")
	}
}

func TestAsTXTRoundTrip(t *testing.T) {
	want := []string{"hello", "v=spf1 -all"}
	rdata, err := resourcerecord.TXTToRData(want)
	if err != nil {
		t.Fatalf("TXTToRData failed: %v", err)
	}
	rr := buildRecord(t, "txt.example.com", llmnr_type.TypeTXT, rdata)

	got, err := rr.AsTXT()
	if err != nil {
		t.Fatalf("AsTXT failed: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("AsTXT returned %d strings, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AsTXT[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	// A character-string longer than 255 bytes cannot be encoded.
	if _, err := resourcerecord.TXTToRData([]string{string(make([]byte, 256))}); err == nil {
		t.Errorf("TXTToRData with 256-byte string should error")
	}
}

func TestAsSRVRoundTrip(t *testing.T) {
	rdata, err := resourcerecord.SRVToRData(10, 20, 389, "dc.example.com")
	if err != nil {
		t.Fatalf("SRVToRData failed: %v", err)
	}
	rr := buildRecord(t, "_ldap._tcp.example.com", llmnr_type.TypeSRV, rdata)

	prio, weight, port, target, err := rr.AsSRV()
	if err != nil {
		t.Fatalf("AsSRV failed: %v", err)
	}
	if prio != 10 || weight != 20 || port != 389 {
		t.Errorf("AsSRV numbers = (%d,%d,%d), want (10,20,389)", prio, weight, port)
	}
	if target != "dc.example.com" {
		t.Errorf("AsSRV target = %q, want %q", target, "dc.example.com")
	}
}

// TestAsPTRCompressedPointer decodes a hand-built PTR record whose PTRDNAME is a
// 0xC0 compression pointer referencing a name earlier in the message. It must be
// resolved through the message-relative path (RFC 1035 §4.1.4). The record is
// decoded via UnmarshalFromMessage at its true offset inside the full buffer.
func TestAsPTRCompressedPointer(t *testing.T) {
	// Lay out a message: at offset 0 sits the target name "host.example.com".
	// A PTR record later in the buffer points its PTRDNAME back to offset 0.
	target := []byte{
		4, 'h', 'o', 's', 't',
		7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		3, 'c', 'o', 'm',
		0,
	}
	msg := append([]byte{}, target...) // target name occupies offsets [0, len(target))
	rrOffset := len(msg)

	// Record NAME: a single label "1" then root, to keep it simple.
	msg = append(msg, 1, '1', 0)
	// TYPE = PTR (12), CLASS = IN (1), TTL = 30
	msg = append(msg, 0x00, 0x0c, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1e)
	// RDLENGTH = 2 (a bare compression pointer)
	msg = append(msg, 0x00, 0x02)
	// RDATA: 0xC000 -> pointer to offset 0
	msg = append(msg, 0xc0, 0x00)

	var rr resourcerecord.ResourceRecord
	n, err := rr.UnmarshalFromMessage(msg, rrOffset)
	if err != nil {
		t.Fatalf("UnmarshalFromMessage failed: %v", err)
	}
	if rrOffset+n != len(msg) {
		t.Fatalf("consumed to %d, want %d", rrOffset+n, len(msg))
	}

	name, err := rr.AsPTR()
	if err != nil {
		t.Fatalf("AsPTR failed: %v", err)
	}
	if name != "host.example.com" {
		t.Errorf("AsPTR (compressed) = %q, want %q", name, "host.example.com")
	}
}

// TestAsSRVCompressedTarget decodes a hand-built SRV record whose target is a
// compression pointer, exercising the same message-relative name resolution for
// the name that follows the SRV priority/weight/port header.
func TestAsSRVCompressedTarget(t *testing.T) {
	target := []byte{
		2, 'd', 'c',
		7, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		3, 'c', 'o', 'm',
		0,
	}
	msg := append([]byte{}, target...)
	rrOffset := len(msg)

	msg = append(msg, 1, 's', 0) // record NAME
	// TYPE = SRV (33), CLASS = IN (1), TTL = 30
	msg = append(msg, 0x00, 0x21, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1e)
	// RDLENGTH = 8 (priority + weight + port + 2-byte pointer)
	msg = append(msg, 0x00, 0x08)
	// priority=1, weight=2, port=88, target -> pointer to offset 0
	msg = append(msg, 0x00, 0x01, 0x00, 0x02, 0x00, 0x58, 0xc0, 0x00)

	var rr resourcerecord.ResourceRecord
	if _, err := rr.UnmarshalFromMessage(msg, rrOffset); err != nil {
		t.Fatalf("UnmarshalFromMessage failed: %v", err)
	}

	prio, weight, port, tgt, err := rr.AsSRV()
	if err != nil {
		t.Fatalf("AsSRV failed: %v", err)
	}
	if prio != 1 || weight != 2 || port != 88 {
		t.Errorf("AsSRV numbers = (%d,%d,%d), want (1,2,88)", prio, weight, port)
	}
	if tgt != "dc.example.com" {
		t.Errorf("AsSRV target (compressed) = %q, want %q", tgt, "dc.example.com")
	}
}

// TestAsAWireSample decodes a hand-built A record and asserts parse-back to the
// right net.IP.
func TestAsAWireSample(t *testing.T) {
	// NAME "x" root, TYPE A, CLASS IN, TTL 30, RDLENGTH 4, RDATA 10.7.0.10
	msg := []byte{
		1, 'x', 0,
		0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1e,
		0x00, 0x04, 10, 7, 0, 10,
	}
	var rr resourcerecord.ResourceRecord
	if _, err := rr.Unmarshal(msg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	ip, err := rr.AsA()
	if err != nil {
		t.Fatalf("AsA failed: %v", err)
	}
	if !ip.Equal(net.IPv4(10, 7, 0, 10)) {
		t.Errorf("AsA = %v, want 10.7.0.10", ip)
	}
	if !bytes.Equal(ip.To4(), []byte{10, 7, 0, 10}) {
		t.Errorf("AsA bytes = %v, want [10 7 0 10]", ip.To4())
	}
}
