package dnsrecord_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dns/dnsrecord"
)

// TestRecordTypeValues pins the numeric values of the record types the tooling relies on,
// including the Microsoft-specific WINS extension and the empty (ZERO) type.
func TestRecordTypeValues(t *testing.T) {
	tests := []struct {
		name  string
		typ   dnsrecord.RecordType
		value uint16
	}{
		{"DNS_TYPE_ZERO", dnsrecord.DNS_TYPE_ZERO, 0},
		{"DNS_TYPE_A", dnsrecord.DNS_TYPE_A, 1},
		{"DNS_TYPE_NS", dnsrecord.DNS_TYPE_NS, 2},
		{"DNS_TYPE_CNAME", dnsrecord.DNS_TYPE_CNAME, 5},
		{"DNS_TYPE_SOA", dnsrecord.DNS_TYPE_SOA, 6},
		{"DNS_TYPE_PTR", dnsrecord.DNS_TYPE_PTR, 12},
		{"DNS_TYPE_MX", dnsrecord.DNS_TYPE_MX, 15},
		{"DNS_TYPE_AAAA", dnsrecord.DNS_TYPE_AAAA, 28},
		{"DNS_TYPE_SRV", dnsrecord.DNS_TYPE_SRV, 33},
		{"DNS_TYPE_WINS", dnsrecord.DNS_TYPE_WINS, 65281},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint16(tt.typ) != tt.value {
				t.Errorf("%s = %d; want %d", tt.name, uint16(tt.typ), tt.value)
			}
		})
	}
}

// TestRecordTypeString verifies the constant-name rendering and the hexadecimal fallback for
// values not present in the specification table.
func TestRecordTypeString(t *testing.T) {
	if got := dnsrecord.DNS_TYPE_SRV.String(); got != "DNS_TYPE_SRV" {
		t.Errorf("DNS_TYPE_SRV.String() = %q; want %q", got, "DNS_TYPE_SRV")
	}
	if got := dnsrecord.DNS_TYPE_WINS.String(); got != "DNS_TYPE_WINS" {
		t.Errorf("DNS_TYPE_WINS.String() = %q; want %q", got, "DNS_TYPE_WINS")
	}
	// 0x1234 is not defined in the table.
	if got := dnsrecord.RecordType(0x1234).String(); got != "RecordType(0x1234)" {
		t.Errorf("RecordType(0x1234).String() = %q; want %q", got, "RecordType(0x1234)")
	}
}
