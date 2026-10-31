package resourcerecord_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

// TestNameRDataOutsideRDataRejected verifies that the name-bearing RDATA
// accessors only decode a name whose in-place bytes lie inside RDATA. Bytes
// that follow RDLENGTH belong to the next record (or are trailing garbage) and
// must not be decoded as this record's name.
func TestNameRDataOutsideRDataRejected(t *testing.T) {
	tests := []struct {
		name  string
		rdata []byte // RDLENGTH and RDATA as they appear on the wire
		trail []byte // bytes after RDATA (e.g. the following record)
	}{
		{
			// RDLENGTH = 0: the name starts beyond RDATA.
			name:  "empty rdata",
			rdata: []byte{0x00, 0x00},
			trail: []byte{4, 'e', 'v', 'i', 'l', 0},
		},
		{
			// RDLENGTH = 2: the name starts inside RDATA but its labels
			// extend past the end of RDATA.
			name:  "name overruns rdata",
			rdata: []byte{0x00, 0x02, 4, 'e'},
			trail: []byte{'v', 'i', 'l', 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := []byte{0} // record NAME: root
			// TYPE = PTR (12), CLASS = IN (1), TTL = 30
			msg = append(msg, 0x00, 0x0c, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1e)
			msg = append(msg, tt.rdata...)
			msg = append(msg, tt.trail...)

			var rr resourcerecord.ResourceRecord
			if _, err := rr.UnmarshalFromMessage(msg, 0); err != nil {
				t.Fatalf("UnmarshalFromMessage failed: %v", err)
			}

			if name, err := rr.AsPTR(); err == nil {
				t.Errorf("AsPTR = %q, want error for a name outside RDATA", name)
			}
		})
	}
}

// TestSRVTargetOutsideRDataRejected verifies the same bound for the SRV target,
// which starts six bytes into RDATA after the priority/weight/port header.
func TestSRVTargetOutsideRDataRejected(t *testing.T) {
	msg := []byte{0} // record NAME: root
	// TYPE = SRV (33), CLASS = IN (1), TTL = 30
	msg = append(msg, 0x00, 0x21, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1e)
	// RDLENGTH = 6: priority=1, weight=2, port=88 and no target
	msg = append(msg, 0x00, 0x06, 0x00, 0x01, 0x00, 0x02, 0x00, 0x58)
	// Following bytes that must not be read as the target
	msg = append(msg, 2, 'd', 'c', 0)

	var rr resourcerecord.ResourceRecord
	if _, err := rr.UnmarshalFromMessage(msg, 0); err != nil {
		t.Fatalf("UnmarshalFromMessage failed: %v", err)
	}

	if _, _, _, tgt, err := rr.AsSRV(); err == nil {
		t.Errorf("AsSRV target = %q, want error for a target outside RDATA", tgt)
	}
}
