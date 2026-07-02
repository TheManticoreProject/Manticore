package msdrsr

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// ---- dcinfo_test.go ----
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

// ---- drs_extensions_test.go ----
// TestExtensionsIntMarshalLayout pins the [MS-DRSR] 5.39 rgb wire layout (verified
// against a live DC): a 52-byte block beginning at dwFlags, with NO leading cb (the cb
// is the DRS_EXTENSIONS.Cb field, not part of rgb).
func TestExtensionsIntMarshalLayout(t *testing.T) {
	e := &DRS_EXTENSIONS_INT{DwFlags: 0x05C08040, Pid: 1234, DwExtCaps: 0xFFFFFFFF}
	b := e.Marshal()
	if len(b) != 52 {
		t.Fatalf("marshalled length = %d, want 52", len(b))
	}
	if got := binary.LittleEndian.Uint32(b[0:4]); got != 0x05C08040 {
		t.Errorf("dwFlags = 0x%08x, want 0x05C08040", got)
	}
	if got := int32(binary.LittleEndian.Uint32(b[20:24])); got != 1234 {
		t.Errorf("Pid = %d, want 1234", got)
	}
	if got := binary.LittleEndian.Uint32(b[48:52]); got != 0xFFFFFFFF {
		t.Errorf("dwExtCaps = 0x%08x, want 0xFFFFFFFF", got)
	}
}

// TestExtensionsIntRoundTrip checks Marshal -> ParseExtensionsInt recovers every field.
func TestExtensionsIntRoundTrip(t *testing.T) {
	in := &DRS_EXTENSIONS_INT{
		Cb:          extIntFieldsSize,
		DwFlags:     DRS_EXT_GETCHGREQ_V8 | DRS_EXT_STRONG_ENCRYPTION,
		Pid:         -7,
		DwReplEpoch: 3,
		DwFlagsExt:  DRS_EXT_RECYCLE_BIN,
		DwExtCaps:   0xDEADBEEF,
	}
	copy(in.SiteObjGuid[:], bytes.Repeat([]byte{0xAB}, 16))
	copy(in.ConfigObjGUID[:], bytes.Repeat([]byte{0xCD}, 16))

	out, err := ParseExtensionsInt(in.Marshal())
	if err != nil {
		t.Fatalf("ParseExtensionsInt: %v", err)
	}
	// Cb is normalised to the spec value on marshal, so compare on that basis.
	in.Cb = extIntFieldsSize
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round trip mismatch:\n in = %+v\nout = %+v", in, out)
	}
}

// TestDefaultClientExtensions asserts the DCSync bind flag set is exactly what the spec
// and the reference client require, so a regression in the constant OR is caught.
func TestDefaultClientExtensions(t *testing.T) {
	e := DefaultClientExtensions()
	const want = DRS_EXT_GETCHGREQ_V6 | DRS_EXT_GETCHGREPLY_V6 | DRS_EXT_GETCHGREQ_V8 |
		DRS_EXT_RESTORE_USN_OPTIMIZATION | DRS_EXT_STRONG_ENCRYPTION | DRS_EXT_NONDOMAIN_NCS
	if e.DwFlags != want {
		t.Errorf("DwFlags = 0x%08x, want 0x%08x", e.DwFlags, want)
	}
	if e.DwFlags&DRS_EXT_STRONG_ENCRYPTION == 0 {
		t.Error("STRONG_ENCRYPTION must be set or the DC will not replicate secrets")
	}
	if e.DwExtCaps != 0xFFFFFFFF {
		t.Errorf("DwExtCaps = 0x%08x, want 0xFFFFFFFF", e.DwExtCaps)
	}
}

// TestExtensionsNDRRoundTrip checks the DRS_EXTENSIONS wrapper marshals as an inline
// conformant byte array (Cb bytes) and round-trips through the NDR codec.
func TestExtensionsNDRRoundTrip(t *testing.T) {
	ext := DefaultClientExtensions().ToExtensions()
	if int(ext.Cb) != len(ext.Rgb) {
		t.Fatalf("Cb %d != len(Rgb) %d", ext.Cb, len(ext.Rgb))
	}
	raw, err := ndr.Marshal(&ext)
	if err != nil {
		t.Fatalf("ndr.Marshal: %v", err)
	}
	var got DRS_EXTENSIONS
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("ndr.Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(ext, got) {
		t.Errorf("NDR round trip mismatch:\n in = %+v\nout = %+v", ext, got)
	}
}

// TestDRSHandleNDRRoundTrip checks the context handle marshals as its 20 octets. A
// context handle is always a struct field (never a top-level parameter), so it is
// exercised here inside a wrapper, mirroring iDL_DRSUnbindRequest.
func TestDRSHandleNDRRoundTrip(t *testing.T) {
	type wrapper struct{ Handle DRS_HANDLE }
	var w wrapper
	for i := range w.Handle {
		w.Handle[i] = byte(i + 1)
	}
	raw, err := ndr.Marshal(&w)
	if err != nil {
		t.Fatalf("ndr.Marshal: %v", err)
	}
	if len(raw) != DRSHandleSize {
		t.Fatalf("marshalled handle length = %d, want %d", len(raw), DRSHandleSize)
	}
	var got wrapper
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("ndr.Unmarshal: %v", err)
	}
	if got.Handle != w.Handle {
		t.Errorf("handle round trip mismatch: in=%v out=%v", w.Handle, got.Handle)
	}
	if w.Handle.IsNull() {
		t.Error("non-zero handle reported as null")
	}
}

// ---- dsname_test.go ----
// TestUUIDWireForm checks UUID holds the 16-octet RPC wire form (guid.GUID.ToBytes
// order) and round-trips back to the same guid.GUID — the encoding the trailing-uint64
// guid.GUID layout would corrupt if used directly under NDR.
func TestUUIDWireForm(t *testing.T) {
	g, err := guid.FromFormatD("e24d201a-4fd6-11d1-a3da-0000f875ae0d")
	if err != nil {
		t.Fatalf("parse guid: %v", err)
	}
	u := UUIDFromGUID(*g)
	if len(u.Octets) != 16 {
		t.Fatalf("UUID is %d octets, want 16", len(u.Octets))
	}
	// Data1 little-endian (e24d201a -> 1a 20 4d e2), Data4 big-endian (a3 da ...).
	want := []byte{0x1a, 0x20, 0x4d, 0xe2, 0xd6, 0x4f, 0xd1, 0x11, 0xa3, 0xda, 0x00, 0x00, 0xf8, 0x75, 0xae, 0x0d}
	for i, b := range want {
		if u.Octets[i] != b {
			t.Fatalf("octet %d = 0x%02x, want 0x%02x (wire order wrong)", i, u.Octets[i], b)
		}
	}
	back := u.GUID()
	if back.ToFormatD() != g.ToFormatD() {
		t.Errorf("round trip: %s != %s", back.ToFormatD(), g.ToFormatD())
	}
}

// TestDSNameGUIDRoundTrip checks a GUID-addressed DSNAME (EXOP_REPL_OBJ form) round-trips
// through NDR behind a unique pointer — exercising the referent id, the hoisted
// maximum_count of the trailing conformant StringName, and the 16-octet GUID.
func TestDSNameGUIDRoundTrip(t *testing.T) {
	type wrapper struct {
		PNC *DSNAME `ndr:"unique"`
	}
	g, _ := guid.FromFormatD("00112233-4455-6677-8899-aabbccddeeff")
	in := NewDSNameFromGUID(*g)
	w := wrapper{PNC: &in}

	raw, err := ndr.Marshal(&w)
	if err != nil {
		t.Fatalf("ndr.Marshal: %v", err)
	}
	var got wrapper
	if err := ndr.Unmarshal(raw, &got); err != nil {
		t.Fatalf("ndr.Unmarshal: %v", err)
	}
	if got.PNC == nil {
		t.Fatal("PNC came back nil")
	}
	if !reflect.DeepEqual(in, *got.PNC) {
		t.Errorf("DSNAME round trip mismatch:\n in = %+v\nout = %+v", in, *got.PNC)
	}
	outGUID := got.PNC.Guid.GUID()
	if outGUID.ToFormatD() != g.ToFormatD() {
		t.Errorf("GUID corrupted: %s != %s", outGUID.ToFormatD(), g.ToFormatD())
	}
}

// TestNewDSNameFromDN checks the DN-addressed DSNAME used for full-NC replication.
func TestNewDSNameFromDN(t *testing.T) {
	dn := "DC=lab,DC=local"
	d := NewDSNameFromDN(dn)
	if int(d.NameLen) != len(dn) {
		t.Errorf("NameLen = %d, want %d", d.NameLen, len(dn))
	}
	if len(d.StringName) != len(dn)+1 || d.StringName[len(dn)] != 0 {
		t.Errorf("StringName length/terminator wrong: len=%d", len(d.StringName))
	}
	if int(d.StructLen) != 56+2*(len(dn)+1) {
		t.Errorf("StructLen = %d, want %d", d.StructLen, 56+2*(len(dn)+1))
	}
	if !d.Guid.IsZero() {
		t.Error("Guid should be zero for a DN-addressed DSNAME")
	}
	if decodeWCharsForTest(d.StringName) != dn {
		t.Errorf("decoded name = %q, want %q", decodeWCharsForTest(d.StringName), dn)
	}
}

func decodeWCharsForTest(u []uint16) string {
	for i, c := range u {
		if c == 0 {
			u = u[:i]
			break
		}
	}
	return string(utf16.Decode(u))
}

// TestNewDSNameFromSID checks the SID-addressed DSNAME used for reverse membership.
func TestNewDSNameFromSID(t *testing.T) {
	sid := []byte{0x01, 0x05, 0, 0, 0, 0, 0, 0x05, 0x15, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 0xf4, 0x01, 0, 0}
	d := NewDSNameFromSID(sid)
	if int(d.SidLen) != len(sid) {
		t.Errorf("SidLen = %d, want %d", d.SidLen, len(sid))
	}
	for i := range sid {
		if d.Sid.Data[i] != sid[i] {
			t.Fatalf("Sid byte %d = 0x%02x, want 0x%02x", i, d.Sid.Data[i], sid[i])
		}
	}
	if !d.Guid.IsZero() {
		t.Error("Guid should be zero for a SID-addressed DSNAME")
	}
}

// ---- getchgreq_test.go ----
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
