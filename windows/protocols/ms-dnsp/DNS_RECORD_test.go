package msdnsp_test

import (
	"bytes"
	"net"
	"testing"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// TestDNSRecordHeaderEndianness pins the mixed-endian header layout: DataLength, Type, Flags,
// Serial, and TimeStamp are little-endian, while TtlSeconds is big-endian per [MS-DNSP]
// section 2.3.2.2. This is the field most likely to be marshaled incorrectly.
func TestDNSRecordHeaderEndianness(t *testing.T) {
	a := msdnsp.NewDNS_RPC_RECORD_A()
	if err := a.SetIPv4(net.ParseIP("192.0.2.1")); err != nil {
		t.Fatalf("SetIPv4 failed: %v", err)
	}

	rec := msdnsp.NewDNS_RECORD()
	rec.Type = msdnsp.DNS_TYPE_A
	rec.Rank = 0xF0
	rec.Serial = 0x11223344
	rec.TtlSeconds = 0x00000E10 // 3600
	rec.TimeStamp = 0xAABBCCDD
	if err := rec.SetData(a); err != nil {
		t.Fatalf("SetData failed: %v", err)
	}

	marshalled, err := rec.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	want := []byte{
		0x04, 0x00, // DataLength = 4, little-endian
		0x01, 0x00, // Type = DNS_TYPE_A (1), little-endian
		0x05,       // Version = 5
		0xF0,       // Rank
		0x00, 0x00, // Flags = 0
		0x44, 0x33, 0x22, 0x11, // Serial, little-endian
		0x00, 0x00, 0x0E, 0x10, // TtlSeconds = 3600, BIG-endian
		0x00, 0x00, 0x00, 0x00, // Reserved = 0
		0xDD, 0xCC, 0xBB, 0xAA, // TimeStamp, little-endian
		192, 0, 2, 1, // Data: A record
	}
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal =\n % x\nwant\n % x", marshalled, want)
	}
}

// TestDNSRecordRoundTripWithA round-trips a full A record through DNS_RECORD and its payload.
func TestDNSRecordRoundTripWithA(t *testing.T) {
	a := msdnsp.NewDNS_RPC_RECORD_A()
	if err := a.SetIPv4(net.ParseIP("203.0.113.9")); err != nil {
		t.Fatalf("SetIPv4 failed: %v", err)
	}
	rec := msdnsp.NewDNS_RECORD()
	rec.Type = msdnsp.DNS_TYPE_A
	rec.TtlSeconds = 600
	if err := rec.SetData(a); err != nil {
		t.Fatalf("SetData failed: %v", err)
	}

	marshalled, err := rec.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := msdnsp.NewDNS_RECORD()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(marshalled) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(marshalled))
	}
	if in.Type != msdnsp.DNS_TYPE_A {
		t.Errorf("Type = %s; want DNS_TYPE_A", in.Type)
	}
	if in.Version != 0x05 {
		t.Errorf("Version = %#x; want 0x05", in.Version)
	}
	if in.TtlSeconds != 600 {
		t.Errorf("TtlSeconds = %d; want 600", in.TtlSeconds)
	}
	if in.DataLength != 4 {
		t.Errorf("DataLength = %d; want 4", in.DataLength)
	}

	parsed := msdnsp.NewDNS_RPC_RECORD_A()
	if _, err := parsed.Unmarshal(in.Data); err != nil {
		t.Fatalf("parsing A payload failed: %v", err)
	}
	if got := parsed.GetIPv4().String(); got != "203.0.113.9" {
		t.Errorf("recovered A = %q; want 203.0.113.9", got)
	}
}

// TestDNSRecordRoundTripWithSOA exercises a variable-length payload (SOA) inside DNS_RECORD to
// confirm DataLength is computed and honored across the header/payload boundary.
func TestDNSRecordRoundTripWithSOA(t *testing.T) {
	soa := msdnsp.NewDNS_RPC_RECORD_SOA()
	soa.DwSerialNo = 2023111401
	soa.DwRefresh = 900
	soa.DwRetry = 600
	soa.DwExpire = 86400
	soa.DwMinimumTtl = 3600
	if err := soa.NamePrimaryServer.SetFQDN("ns1.example.com"); err != nil {
		t.Fatalf("SetFQDN(primary) failed: %v", err)
	}
	if err := soa.ZoneAdminEmail.SetFQDN("hostmaster.example.com"); err != nil {
		t.Fatalf("SetFQDN(admin) failed: %v", err)
	}

	rec := msdnsp.NewDNS_RECORD()
	rec.Type = msdnsp.DNS_TYPE_SOA
	if err := rec.SetData(soa); err != nil {
		t.Fatalf("SetData failed: %v", err)
	}

	marshalled, err := rec.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := msdnsp.NewDNS_RECORD()
	if _, err := in.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if int(in.DataLength) != len(in.Data) {
		t.Errorf("DataLength %d != len(Data) %d", in.DataLength, len(in.Data))
	}

	parsed := msdnsp.NewDNS_RPC_RECORD_SOA()
	if _, err := parsed.Unmarshal(in.Data); err != nil {
		t.Fatalf("parsing SOA payload failed: %v", err)
	}
	if parsed.DwSerialNo != 2023111401 {
		t.Errorf("recovered SOA serial = %d; want 2023111401", parsed.DwSerialNo)
	}
	primary, _ := parsed.NamePrimaryServer.GetFQDN()
	if primary != "ns1.example.com" {
		t.Errorf("recovered primary = %q; want ns1.example.com", primary)
	}
}

// TestDNSRecordUnmarshalTruncated verifies both a short header and a Data field shorter than
// the declared DataLength are rejected.
func TestDNSRecordUnmarshalTruncated(t *testing.T) {
	in := msdnsp.NewDNS_RECORD()
	if _, err := in.Unmarshal(make([]byte, 23)); err == nil {
		t.Errorf("expected error for short header, got nil")
	}

	// Header declares DataLength=8 but only 2 data bytes follow.
	header := make([]byte, 24)
	header[0] = 0x08 // DataLength low byte
	truncated := append(header, 0x00, 0x00)
	if _, err := in.Unmarshal(truncated); err == nil {
		t.Errorf("expected error for truncated Data, got nil")
	}
}
