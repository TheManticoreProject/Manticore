package dnsrecord_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dns/dnsrecord"
)

// TestRecordNamePreferenceWireShape verifies the big-endian wPreference followed by the
// DNS_COUNT_NAME exchange name (used by MX, AFSDB, RT).
func TestRecordNamePreferenceWireShape(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_NAME_PREFERENCE()
	r.WPreference = 0x000A // 10
	if err := r.NameExchange.SetFQDN("mail.example.com"); err != nil {
		t.Fatalf("SetFQDN failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// Big-endian preference: 00 0A.
	if !bytes.Equal(marshalled[:2], []byte{0x00, 0x0A}) {
		t.Errorf("wPreference = % x; want 00 0A", marshalled[:2])
	}
}

// TestRecordNamePreferenceRoundTrip round-trips an MX-style record.
func TestRecordNamePreferenceRoundTrip(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_NAME_PREFERENCE()
	r.WPreference = 20
	if err := r.NameExchange.SetFQDN("mx2.example.com"); err != nil {
		t.Fatalf("SetFQDN failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := dnsrecord.NewDNS_RPC_RECORD_NAME_PREFERENCE()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(marshalled) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(marshalled))
	}
	if in.WPreference != 20 {
		t.Errorf("WPreference = %d; want 20", in.WPreference)
	}
	exchange, _ := in.NameExchange.GetFQDN()
	if exchange != "mx2.example.com" {
		t.Errorf("NameExchange = %q; want mx2.example.com", exchange)
	}
}
