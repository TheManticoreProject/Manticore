package ndr

import (
	"bytes"
	"reflect"
	"testing"
)

// efsBlob models EFS_RPC_BLOB { DWORD cbData; [size_is(cbData)] byte* bData; }.
type efsBlob struct {
	CbData DWORD  `ndr:"dword"`
	BData  []byte `ndr:"unique,conformant,size_is=CbData"`
}

// efsCall models an EFSR-style call: scalar, unique pointer to struct, wide string
// pointer, scalar.
type efsCall struct {
	Flags     DWORD    `ndr:"dword"`
	Reserved  *efsBlob `ndr:"unique"`
	FileName  WSTR     `ndr:"unique"`
	InfoClass DWORD    `ndr:"dword"`
}

func (*efsCall) Opnum() uint16 { return 16 }

func TestMarshal_DeferredOrdering_GoldenBytes(t *testing.T) {
	call := &efsCall{
		Flags:     0x11,
		Reserved:  &efsBlob{CbData: 2, BData: []byte{0xaa, 0xbb}},
		FileName:  "A",
		InfoClass: 0x22,
	}
	got, err := Marshal(call)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	want := []byte{
		// --- top-level inline ---
		0x11, 0x00, 0x00, 0x00, // Flags
		0x00, 0x00, 0x02, 0x00, // Reserved referent id (0x00020000)
		0x04, 0x00, 0x02, 0x00, // FileName referent id (0x00020004)
		0x22, 0x00, 0x00, 0x00, // InfoClass
		// --- deferred[0]: Reserved (efsBlob) construction ---
		0x02, 0x00, 0x00, 0x00, // CbData
		0x08, 0x00, 0x02, 0x00, // BData referent id (0x00020008)
		0x02, 0x00, 0x00, 0x00, // BData conformant maximum_count
		0xaa, 0xbb, // BData elements
		// --- deferred[1]: FileName wstring ("A") ---
		0x00, 0x00, // 2-byte pad to 4-align the count
		0x02, 0x00, 0x00, 0x00, // maximum_count (incl NUL)
		0x00, 0x00, 0x00, 0x00, // offset
		0x02, 0x00, 0x00, 0x00, // actual_count
		0x41, 0x00, // "A"
		0x00, 0x00, // terminator
	}
	if !bytes.Equal(got, want) {
		t.Errorf("EFSR-style marshal:\n got %x\nwant %x", got, want)
	}
}

func TestMarshalUnmarshal_RoundTrip(t *testing.T) {
	in := &efsCall{
		Flags:     0xdeadbeef,
		Reserved:  &efsBlob{CbData: 3, BData: []byte{1, 2, 3}},
		FileName:  `\\host\share`,
		InfoClass: 7,
	}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out efsCall
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, &out) {
		t.Errorf("round trip mismatch:\n in  %+v (blob %+v)\n out %+v (blob %+v)", in, in.Reserved, &out, out.Reserved)
	}
}

func TestMarshal_NullPointer(t *testing.T) {
	call := &efsCall{Flags: 1, Reserved: nil, FileName: "x", InfoClass: 2}
	raw, err := Marshal(call)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Reserved is the second field: its referent id must be 0 (NULL).
	if !bytes.Equal(raw[4:8], []byte{0, 0, 0, 0}) {
		t.Errorf("NULL Reserved referent = %x, want 00000000", raw[4:8])
	}

	var out efsCall
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Reserved != nil {
		t.Errorf("Reserved = %+v, want nil after a NULL referent", out.Reserved)
	}
}

func TestRequest_UsesOpnumlessMarshalling(t *testing.T) {
	// Request marshals only the struct's fields (Opnum is a method, not a field).
	call := &efsCall{Flags: 0x11, FileName: "A", InfoClass: 0x22}
	raw, err := Request(call)
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if len(raw) < 4 || !bytes.Equal(raw[0:4], []byte{0x11, 0x00, 0x00, 0x00}) {
		t.Errorf("Request stub does not start with Flags: %x", raw)
	}
	if call.Opnum() != 16 {
		t.Errorf("Opnum = %d, want 16", call.Opnum())
	}
}

func TestScalars_RoundTrip(t *testing.T) {
	type scalars struct {
		A bool   `ndr:"bool"`
		B uint8  `ndr:"byte"`
		C uint16 `ndr:"word"`
		D uint32 `ndr:"dword"`
		E uint64 `ndr:"hyper"`
		F int32  `ndr:"long"`
	}
	in := &scalars{A: true, B: 0x7f, C: 0xbeef, D: 0xdeadbeef, E: 0x1122334455667788, F: -5}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out scalars
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if *in != out {
		t.Errorf("round trip: in %+v out %+v", in, out)
	}
}

// fixedThing is a custom NDR type using the Marshaler escape hatch: a fixed 4-byte
// value with 4-byte alignment.
type fixedThing struct{ V uint32 }

func (f *fixedThing) AlignmentNDR() int { return 4 }
func (f *fixedThing) MarshalNDR(e *Encoder) error {
	e.WriteUint32(f.V)
	return nil
}
func (f *fixedThing) UnmarshalNDR(d *Decoder) error {
	v, err := d.ReadUint32()
	f.V = v
	return err
}

func TestMarshaler_EscapeHatch(t *testing.T) {
	type wrap struct {
		Lead uint8
		T    fixedThing
	}
	in := &wrap{Lead: 0x01, T: fixedThing{V: 0xcafebabe}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// uint8 then 4-aligned custom uint32.
	want := []byte{0x01, 0x00, 0x00, 0x00, 0xbe, 0xba, 0xfe, 0xca}
	if !bytes.Equal(raw, want) {
		t.Errorf("escape hatch:\n got %x\nwant %x", raw, want)
	}
	var out wrap
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != *in {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}

// TestBOOL_IsFourOctets verifies the Windows BOOL alias encodes as a 4-octet
// integer, not a 1-octet NDR boolean (issue #393).
func TestBOOL_IsFourOctets(t *testing.T) {
	type withBool struct {
		Lead uint8
		Flag BOOL // Windows BOOL: 4 octets
	}
	raw, err := Marshal(&withBool{Lead: 0x01, Flag: 1})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// uint8 then a 4-aligned 4-octet BOOL value of 1.
	want := []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}
	if !bytes.Equal(raw, want) {
		t.Errorf("BOOL encoding:\n got %x\nwant %x", raw, want)
	}

	var out withBool
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Flag != 1 || out.Lead != 1 {
		t.Errorf("round trip: got %+v", out)
	}
}

// TestBOOLEAN_IsOneOctet verifies the NDR boolean (Go bool) stays a single octet.
func TestBOOLEAN_IsOneOctet(t *testing.T) {
	type withBoolean struct {
		A BOOLEAN
		B uint8
	}
	raw, err := Marshal(&withBoolean{A: true, B: 0x7f})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{0x01, 0x7f}
	if !bytes.Equal(raw, want) {
		t.Errorf("BOOLEAN encoding:\n got %x\nwant %x", raw, want)
	}
}

// TestAlignTag_Applied verifies the align=N struct tag inserts padding (issue #394).
func TestAlignTag_Applied(t *testing.T) {
	type withAlign struct {
		A uint8
		B uint32 `ndr:"align=8"`
	}
	raw, err := Marshal(&withAlign{A: 1, B: 2})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// A at offset 0; align=8 pads to offset 8; B (4 octets) at offset 8.
	want := []byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0x02, 0, 0, 0}
	if !bytes.Equal(raw, want) {
		t.Errorf("align=8:\n got %x\nwant %x", raw, want)
	}

	var out withAlign
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != (withAlign{A: 1, B: 2}) {
		t.Errorf("round trip: got %+v", out)
	}
}

// TestEmbeddedConformantArray_Hoisted verifies that a conformant array embedded
// directly in a struct has its maximum_count hoisted to the start (issue #395).
func TestEmbeddedConformantArray_Hoisted(t *testing.T) {
	type rec struct {
		N    uint32
		Data []byte `ndr:"conformant"`
	}
	in := &rec{N: 0xAABBCCDD, Data: []byte{0x01, 0x02, 0x03}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x03, 0x00, 0x00, 0x00, // hoisted maximum_count
		0xDD, 0xCC, 0xBB, 0xAA, // N
		0x01, 0x02, 0x03, // elements, in place at the end
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("hoisted conformant array:\n got %x\nwant %x", raw, want)
	}

	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.N != in.N || !bytes.Equal(out.Data, in.Data) {
		t.Errorf("round trip: got %+v want %+v", out, in)
	}
}
