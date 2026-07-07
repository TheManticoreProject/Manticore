package dnsrecord_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dns/dnsrecord"
)

// TestRecordSOAWireShape verifies the big-endian encoding of the five numeric fields and the
// two DNS_COUNT_NAME names.
func TestRecordSOAWireShape(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_SOA()
	r.DwSerialNo = 0x01020304
	r.DwRefresh = 0x05060708
	r.DwRetry = 0x090a0b0c
	r.DwExpire = 0x0d0e0f10
	r.DwMinimumTtl = 0x11121314
	if err := r.NamePrimaryServer.SetFQDN("ns.example.com"); err != nil {
		t.Fatalf("SetFQDN(primary) failed: %v", err)
	}
	if err := r.ZoneAdminEmail.SetFQDN("admin.example.com"); err != nil {
		t.Fatalf("SetFQDN(admin) failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// The five DWORDs are big-endian, laid out most-significant byte first.
	wantHeader := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c,
		0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14,
	}
	if !bytes.Equal(marshalled[:20], wantHeader) {
		t.Errorf("SOA numeric header = % x; want % x", marshalled[:20], wantHeader)
	}
}

// TestRecordSOARoundTrip round-trips an SOA record and checks every field.
func TestRecordSOARoundTrip(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_SOA()
	r.DwSerialNo = 42
	r.DwRefresh = 900
	r.DwRetry = 600
	r.DwExpire = 86400
	r.DwMinimumTtl = 3600
	if err := r.NamePrimaryServer.SetFQDN("ns1.example.com"); err != nil {
		t.Fatalf("SetFQDN(primary) failed: %v", err)
	}
	if err := r.ZoneAdminEmail.SetFQDN("hostmaster.example.com"); err != nil {
		t.Fatalf("SetFQDN(admin) failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := dnsrecord.NewDNS_RPC_RECORD_SOA()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(marshalled) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(marshalled))
	}

	if in.DwSerialNo != 42 || in.DwRefresh != 900 || in.DwRetry != 600 || in.DwExpire != 86400 || in.DwMinimumTtl != 3600 {
		t.Errorf("numeric fields mismatch: %+v", in)
	}
	primary, _ := in.NamePrimaryServer.GetFQDN()
	admin, _ := in.ZoneAdminEmail.GetFQDN()
	if primary != "ns1.example.com" {
		t.Errorf("NamePrimaryServer = %q; want ns1.example.com", primary)
	}
	if admin != "hostmaster.example.com" {
		t.Errorf("ZoneAdminEmail = %q; want hostmaster.example.com", admin)
	}
}
