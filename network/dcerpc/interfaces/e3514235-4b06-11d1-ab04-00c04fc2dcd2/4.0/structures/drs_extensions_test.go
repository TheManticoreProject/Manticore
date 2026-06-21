package structures

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestExtensionsIntMarshalLayout pins the [MS-DRSR] 5.39 wire layout: a 56-byte blob
// whose Cb field counts the 52 bytes that follow it, with dwFlags immediately after Cb.
func TestExtensionsIntMarshalLayout(t *testing.T) {
	e := &DRS_EXTENSIONS_INT{DwFlags: 0x05C08040, Pid: 1234, DwExtCaps: 0xFFFFFFFF}
	b := e.Marshal()
	if len(b) != 56 {
		t.Fatalf("marshalled length = %d, want 56", len(b))
	}
	if got := binary.LittleEndian.Uint32(b[0:4]); got != extIntFieldsSize {
		t.Errorf("Cb = %d, want %d", got, extIntFieldsSize)
	}
	if got := binary.LittleEndian.Uint32(b[4:8]); got != 0x05C08040 {
		t.Errorf("dwFlags = 0x%08x, want 0x05C08040", got)
	}
	if got := int32(binary.LittleEndian.Uint32(b[24:28])); got != 1234 {
		t.Errorf("Pid = %d, want 1234", got)
	}
	if got := binary.LittleEndian.Uint32(b[52:56]); got != 0xFFFFFFFF {
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
