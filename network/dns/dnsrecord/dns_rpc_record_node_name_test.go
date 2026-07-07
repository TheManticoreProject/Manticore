package dnsrecord_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dns/dnsrecord"
)

// TestRecordNodeNameRoundTrip round-trips a NODE_NAME record (used by NS, PTR, CNAME, etc.)
// and verifies the embedded DNS_COUNT_NAME encoding.
func TestRecordNodeNameRoundTrip(t *testing.T) {
	r := dnsrecord.NewDNS_RPC_RECORD_NODE_NAME()
	if err := r.NameNode.SetFQDN("host.example.com"); err != nil {
		t.Fatalf("SetFQDN failed: %v", err)
	}

	marshalled, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// 04 'host' 07 'example' 03 'com' 00, prefixed by Length(21) and LabelCount(3).
	wantRaw := []byte{0x04, 'h', 'o', 's', 't', 0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00}
	want := append([]byte{byte(len(wantRaw)), 3}, wantRaw...)
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal = % x; want % x", marshalled, want)
	}

	in := dnsrecord.NewDNS_RPC_RECORD_NODE_NAME()
	read, err := in.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(marshalled) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(marshalled))
	}
	got, err := in.NameNode.GetFQDN()
	if err != nil {
		t.Fatalf("GetFQDN failed: %v", err)
	}
	if got != "host.example.com" {
		t.Errorf("NameNode = %q; want host.example.com", got)
	}
}
