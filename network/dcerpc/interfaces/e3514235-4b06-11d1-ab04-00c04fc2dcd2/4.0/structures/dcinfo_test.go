package structures

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestDCInfoReqV1RoundTrip checks the DCINFOREQ union (switch Tag + V1 arm with a
// [string] domain pointer and InfoLevel) round-trips through NDR.
func TestDCInfoReqV1RoundTrip(t *testing.T) {
	dom := ndr.WSTR("lab.local")
	in := DRS_MSG_DCINFOREQ{
		Tag: 1,
		V1:  DRS_MSG_DCINFOREQ_V1{Domain: &dom, InfoLevel: 2},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got DRS_MSG_DCINFOREQ
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Tag != 1 || got.V1.InfoLevel != 2 {
		t.Fatalf("Tag=%d InfoLevel=%d, want 1/2", got.Tag, got.V1.InfoLevel)
	}
	if got.V1.Domain == nil || string(*got.V1.Domain) != "lab.local" {
		t.Errorf("Domain = %v, want lab.local", got.V1.Domain)
	}
}
