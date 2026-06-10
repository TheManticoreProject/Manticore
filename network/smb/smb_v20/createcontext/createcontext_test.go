package createcontext

import (
	"bytes"
	"testing"
)

func TestMarshalParseSingleNoData(t *testing.T) {
	in := []CreateContext{{Name: NameQueryMaximalAccess}}
	buf, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d contexts, want 1", len(got))
	}
	if !bytes.Equal(got[0].Name, NameQueryMaximalAccess) {
		t.Errorf("name = % x, want % x", got[0].Name, NameQueryMaximalAccess)
	}
	if len(got[0].Data) != 0 {
		t.Errorf("expected no data, got % x", got[0].Data)
	}
}

func TestMarshalParseWithData(t *testing.T) {
	in := []CreateContext{
		{Name: NameDurableHandleReq, Data: bytes.Repeat([]byte{0xAB}, 16)},
		{Name: NameAllocationSize, Data: []byte{0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}
	buf, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := Parse(buf)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("parsed %d contexts, want 2", len(got))
	}
	for i := range in {
		if !bytes.Equal(got[i].Name, in[i].Name) {
			t.Errorf("context %d name = % x, want % x", i, got[i].Name, in[i].Name)
		}
		if !bytes.Equal(got[i].Data, in[i].Data) {
			t.Errorf("context %d data = % x, want % x", i, got[i].Data, in[i].Data)
		}
	}
}

func TestMarshalEmpty(t *testing.T) {
	buf, err := Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal(nil): %v", err)
	}
	if buf != nil {
		t.Errorf("Marshal(nil) = % x, want nil", buf)
	}
}

func TestParseRejectsTruncatedHeader(t *testing.T) {
	if _, err := Parse([]byte{0x00, 0x00, 0x00}); err == nil {
		t.Error("Parse should reject a buffer shorter than the context header")
	}
}

func TestParseRejectsOutOfBoundsNameOffset(t *testing.T) {
	// Build a valid single context, then corrupt its NameLength to run past the
	// buffer; Parse must reject it rather than slicing out of range.
	buf, err := Marshal([]CreateContext{{Name: NameRequestLease, Data: []byte{1, 2, 3, 4}}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// NameLength is at bytes [6:8]; set it absurdly large.
	buf[6] = 0xFF
	buf[7] = 0xFF
	if _, err := Parse(buf); err == nil {
		t.Error("Parse should reject a name length that exceeds the buffer")
	}
}

func TestParseRejectsCyclicNext(t *testing.T) {
	// A context whose Next points back into itself (< headerSize) is rejected.
	buf, err := Marshal([]CreateContext{{Name: NameTimewarpToken}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Set Next (bytes [0:4]) to 4, which is < headerSize.
	buf[0] = 0x04
	if _, err := Parse(buf); err == nil {
		t.Error("Parse should reject a Next offset smaller than the header")
	}
}
