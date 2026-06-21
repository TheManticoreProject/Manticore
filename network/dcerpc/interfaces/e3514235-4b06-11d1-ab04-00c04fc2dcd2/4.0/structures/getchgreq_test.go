package structures

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// TestGetChgReqV8RoundTrip marshals a single-object (EXOP_REPL_OBJ) GETCHGREQ union and
// unmarshals it, confirming the switch Tag selects the V8 arm, the nil unique pointers
// encode as NULL referents, and the embedded GUID-addressed DSNAME survives.
func TestGetChgReqV8RoundTrip(t *testing.T) {
	g, _ := guid.FromFormatD("00112233-4455-6677-8899-aabbccddeeff")
	pnc := NewDSNameFromGUID(*g)
	in := DRS_MSG_GETCHGREQ{
		Tag: 8,
		V8: DRS_MSG_GETCHGREQ_V8{
			PNC:          &pnc,
			UlFlags:      ndr.DWORD(DRS_INIT_SYNC | DRS_WRIT_REP),
			CMaxObjects:  1,
			UlExtendedOp: ndr.DWORD(EXOP_REPL_OBJ),
		},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got DRS_MSG_GETCHGREQ
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Tag != 8 {
		t.Fatalf("Tag = %d, want 8", got.Tag)
	}
	if got.V8.UlExtendedOp != ndr.DWORD(EXOP_REPL_OBJ) {
		t.Errorf("UlExtendedOp = %d, want %d", got.V8.UlExtendedOp, EXOP_REPL_OBJ)
	}
	if got.V8.UlFlags != ndr.DWORD(DRS_INIT_SYNC|DRS_WRIT_REP) {
		t.Errorf("UlFlags = 0x%x, want 0x%x", got.V8.UlFlags, DRS_INIT_SYNC|DRS_WRIT_REP)
	}
	if got.V8.PNC == nil {
		t.Fatal("PNC came back nil")
	}
	outGUID := got.V8.PNC.Guid.GUID()
	if outGUID.ToFormatD() != g.ToFormatD() {
		t.Errorf("pNC GUID corrupted: %s != %s", outGUID.ToFormatD(), g.ToFormatD())
	}
}

// TestCrackReqV1RoundTrip marshals a CRACKREQ union with two names and confirms the V1
// arm, the name count, and the array of [string] pointers round-trip.
func TestCrackReqV1RoundTrip(t *testing.T) {
	n1, n2 := ndr.WSTR(`LAB\\krbtgt`), ndr.WSTR(`LAB\\Administrator`)
	in := DRS_MSG_CRACKREQ{
		Tag: 1,
		V1: DRS_MSG_CRACKREQ_V1{
			FormatOffered: ndr.DWORD(DS_NT4_ACCOUNT_NAME),
			FormatDesired: ndr.DWORD(DS_UNIQUE_ID_NAME),
			CNames:        2,
			RpNames:       []*ndr.WSTR{&n1, &n2},
		},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got DRS_MSG_CRACKREQ
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Tag != 1 || got.V1.CNames != 2 || len(got.V1.RpNames) != 2 {
		t.Fatalf("Tag=%d CNames=%d names=%d, want 1/2/2", got.Tag, got.V1.CNames, len(got.V1.RpNames))
	}
	if got.V1.RpNames[0] == nil || string(*got.V1.RpNames[0]) != string(n1) {
		t.Errorf("name[0] = %v, want %q", got.V1.RpNames[0], string(n1))
	}
	if got.V1.FormatDesired != ndr.DWORD(DS_UNIQUE_ID_NAME) {
		t.Errorf("FormatDesired = %d, want %d", got.V1.FormatDesired, DS_UNIQUE_ID_NAME)
	}
}
