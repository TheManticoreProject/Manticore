package regf

import (
	"bytes"
	"flag"
	"os"
	"reflect"
	"testing"
)

var update = flag.Bool("update", false, "regenerate testdata golden files")

const goldenPath = "testdata/sample.hive"

// TestGoldenFixture keeps the committed sample hive in sync with buildSampleHive. Run
// `go test -run TestGoldenFixture -update` to regenerate testdata/sample.hive after
// changing the builder.
func TestGoldenFixture(t *testing.T) {
	want := buildSampleHive()
	if *update {
		if err := os.WriteFile(goldenPath, want, 0o644); err != nil {
			t.Fatalf("writing golden fixture: %v", err)
		}
		t.Logf("wrote %d bytes to %s", len(want), goldenPath)
	}
	got, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden fixture (run with -update to create it): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("golden fixture %s is out of date; run: go test -run TestGoldenFixture -update", goldenPath)
	}
}

// --- structure round-trips (Marshal -> Unmarshal -> equal) ---

func TestBaseBlockRoundTrip(t *testing.T) {
	orig := BaseBlock{
		Signature:        regfSignature,
		MajorVersion:     1,
		MinorVersion:     5,
		FileFormat:       1,
		RootCellOffset:   0x20,
		HiveBinsDataSize: 4096,
		ClusteringFactor: 1,
		Checksum:         0xDEADBEEF,
	}
	raw, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != baseBlockSize {
		t.Fatalf("marshaled length = %d, want %d", len(raw), baseBlockSize)
	}
	var got BaseBlock
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
}

func TestHiveBinRoundTrip(t *testing.T) {
	orig := HiveBin{Signature: hbinSignature, Offset: 0, Size: 4096, Timestamp: 0x1234}
	raw, _ := orig.Marshal()
	var got HiveBin
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
}

func TestKeyNodeRoundTrip(t *testing.T) {
	orig := KeyNode{
		Signature:         nkSignature,
		Flags:             KeyCompName,
		NumberOfSubKeys:   3,
		SubKeysListOffset: 0x40,
		NumberOfValues:    2,
		ValuesListOffset:  0x80,
		SecurityOffset:    nullCellOffset,
		ClassNameOffset:   nullCellOffset,
		KeyNameLength:     uint16(len("Software")),
		KeyNameRaw:        []byte("Software"),
	}
	raw, _ := orig.Marshal()
	var got KeyNode
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
	if got.Name() != "Software" {
		t.Errorf("Name() = %q, want %q", got.Name(), "Software")
	}
}

func TestKeyValueRoundTrip(t *testing.T) {
	orig := KeyValue{
		Signature:  vkSignature,
		NameLength: uint16(len("StrVal")),
		DataSize:   12,
		DataOffset: 0x100,
		DataType:   RegSz,
		Flags:      ValueCompName,
		NameRaw:    []byte("StrVal"),
	}
	raw, _ := orig.Marshal()
	var got KeyValue
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
}

func TestSubKeyListRoundTrip(t *testing.T) {
	orig := SubKeyList{Signature: lhSig, NumberOfElements: 2, Elements: make([]byte, 16)}
	for i := range orig.Elements {
		orig.Elements[i] = byte(i)
	}
	raw, _ := orig.Marshal()
	var got SubKeyList
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
	if offs := got.KeyNodeOffsets(); len(offs) != 2 {
		t.Errorf("KeyNodeOffsets returned %d offsets, want 2", len(offs))
	}
}

func TestSecurityKeyRoundTrip(t *testing.T) {
	sd := []byte{1, 0, 0, 0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	orig := SecurityKey{
		Signature:              skSignature,
		ReferenceCount:         3,
		SecurityDescriptorSize: uint32(len(sd)),
		SecurityDescriptor:     sd,
	}
	raw, _ := orig.Marshal()
	var got SecurityKey
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
}

func TestBigDataRoundTrip(t *testing.T) {
	orig := BigData{Signature: dbSignature, NumberOfSegments: 4, SegmentsListOffset: 0x200}
	raw, _ := orig.Marshal()
	var got BigData
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
}

func TestUnmarshalRejectsBadSignature(t *testing.T) {
	bad := make([]byte, baseBlockSize)
	var bb BaseBlock
	if _, err := bb.Unmarshal(bad); err == nil {
		t.Error("BaseBlock.Unmarshal accepted a zero (invalid) signature")
	}
	nkBad := make([]byte, keyNodeMinSize)
	var nk KeyNode
	if _, err := nk.Unmarshal(nkBad); err == nil {
		t.Error("KeyNode.Unmarshal accepted a zero (invalid) signature")
	}
}

// --- navigation against the committed fixture (exercises the on-disk Open path) ---

func TestSampleHiveNavigation(t *testing.T) {
	h, err := Open(goldenPath)
	if err != nil {
		t.Fatalf("Open(%s): %v", goldenPath, err)
	}
	defer h.Close()

	root, err := h.RootKey()
	if err != nil {
		t.Fatalf("RootKey: %v", err)
	}
	if root.Name() != "ROOT" {
		t.Errorf("root name = %q, want %q", root.Name(), "ROOT")
	}
	if !root.IsRoot() {
		t.Error("root.IsRoot() = false, want true")
	}

	if keys, err := h.EnumKey(""); err != nil {
		t.Fatalf("EnumKey: %v", err)
	} else if !reflect.DeepEqual(keys, []string{"Sub"}) {
		t.Errorf("EnumKey(\"\") = %v, want [Sub]", keys)
	}

	if _, err := h.FindKey("Sub"); err != nil {
		t.Errorf("FindKey(Sub): %v", err)
	}
	if _, err := h.FindKey("DoesNotExist"); err == nil {
		t.Error("FindKey(DoesNotExist) succeeded, want error")
	}

	values, err := h.EnumValues("Sub")
	if err != nil {
		t.Fatalf("EnumValues(Sub): %v", err)
	}
	if !reflect.DeepEqual(values, []string{"DwordVal", "StrVal", "BigVal"}) {
		t.Errorf("EnumValues(Sub) = %v, want [DwordVal StrVal BigVal]", values)
	}

	// Inline REG_DWORD.
	typ, data, err := h.GetValue("Sub", "DwordVal")
	if err != nil {
		t.Fatalf("GetValue(Sub, DwordVal): %v", err)
	}
	if typ != RegDword {
		t.Errorf("DwordVal type = %d, want REG_DWORD(%d)", typ, RegDword)
	}
	if len(data) != 4 || data[0] != 0x78 || data[1] != 0x56 || data[2] != 0x34 || data[3] != 0x12 {
		t.Errorf("DwordVal data = % x, want 78 56 34 12", data)
	}

	// External REG_SZ.
	typ, _, err = h.GetValue("Sub", "StrVal")
	if err != nil {
		t.Fatalf("GetValue(Sub, StrVal): %v", err)
	}
	if typ != RegSz {
		t.Errorf("StrVal type = %d, want REG_SZ(%d)", typ, RegSz)
	}
	sub, _ := h.FindKey("Sub")
	v, _ := sub.Value("StrVal")
	if got := v.String(); got != "hello" {
		t.Errorf("StrVal string = %q, want %q", got, "hello")
	}
	dv, _ := sub.Value("DwordVal")
	if n, ok := dv.Uint32(); !ok || n != 0x12345678 {
		t.Errorf("DwordVal Uint32 = 0x%X (ok=%v), want 0x12345678", n, ok)
	}

	// Big-data (db) value: 20000 bytes reassembled from two segments.
	typ, data, err = h.GetValue("Sub", "BigVal")
	if err != nil {
		t.Fatalf("GetValue(Sub, BigVal): %v", err)
	}
	if typ != RegBinary {
		t.Errorf("BigVal type = %d, want REG_BINARY(%d)", typ, RegBinary)
	}
	if !bytes.Equal(data, bigValBytes()) {
		t.Errorf("BigVal data mismatch: got %d bytes, want %d bytes (content differs)", len(data), bigValSize)
	}

	// Class data on the root key.
	class, err := h.GetClass("")
	if err != nil {
		t.Fatalf("GetClass(root): %v", err)
	}
	if string(class) != "classdata" {
		t.Errorf("root class = %q, want %q", class, "classdata")
	}

	// SK record: ROOT and Sub share one security descriptor.
	for _, key := range []string{"", "Sub"} {
		sd, err := h.GetSecurity(key)
		if err != nil {
			t.Fatalf("GetSecurity(%q): %v", key, err)
		}
		if len(sd) != 20 {
			t.Errorf("GetSecurity(%q) returned %d bytes, want 20", key, len(sd))
		}
		if sd[0] != 1 {
			t.Errorf("GetSecurity(%q) revision = %d, want 1", key, sd[0])
		}
	}
}

func TestOpenBytesRejectsGarbage(t *testing.T) {
	if _, err := OpenBytes(make([]byte, 10)); err == nil {
		t.Error("OpenBytes accepted a 10-byte slice, want error")
	}
	if _, err := OpenBytes(make([]byte, baseBlockSize)); err == nil {
		t.Error("OpenBytes accepted a base block with an invalid signature, want error")
	}
}
