package attributes

import (
	"encoding/binary"
	"testing"
)

func TestManagedPasswordBlobRoundTrip(t *testing.T) {
	blob := &ManagedPasswordBlob{
		CurrentPassword:                "P@ssw0rd!",
		PreviousPassword:               "OldP@ss",
		QueryPasswordIntervalTicks:     864000000000,    // 1 day in 100ns ticks
		UnchangedPasswordIntervalTicks: 2592000000000,   // 30 days
	}

	wire := MarshalManagedPasswordBlob(blob)

	// Check version.
	if v := binary.LittleEndian.Uint16(wire[0:2]); v != 1 {
		t.Fatalf("version = %d, want 1", v)
	}

	// Check length field matches actual size.
	if l := binary.LittleEndian.Uint32(wire[4:8]); int(l) != len(wire) {
		t.Fatalf("length field = %d, wire length = %d", l, len(wire))
	}

	parsed, err := ParseManagedPasswordBlob(wire)
	if err != nil {
		t.Fatalf("ParseManagedPasswordBlob: %v", err)
	}

	if parsed.CurrentPassword != blob.CurrentPassword {
		t.Errorf("CurrentPassword = %q, want %q", parsed.CurrentPassword, blob.CurrentPassword)
	}
	if parsed.PreviousPassword != blob.PreviousPassword {
		t.Errorf("PreviousPassword = %q, want %q", parsed.PreviousPassword, blob.PreviousPassword)
	}
	if parsed.QueryPasswordIntervalTicks != blob.QueryPasswordIntervalTicks {
		t.Errorf("QueryPasswordIntervalTicks = %d, want %d", parsed.QueryPasswordIntervalTicks, blob.QueryPasswordIntervalTicks)
	}
	if parsed.UnchangedPasswordIntervalTicks != blob.UnchangedPasswordIntervalTicks {
		t.Errorf("UnchangedPasswordIntervalTicks = %d, want %d", parsed.UnchangedPasswordIntervalTicks, blob.UnchangedPasswordIntervalTicks)
	}
}

func TestManagedPasswordBlobNoPreviousPassword(t *testing.T) {
	blob := &ManagedPasswordBlob{
		CurrentPassword:                "OnlyCurrentPwd",
		QueryPasswordIntervalTicks:     100000000,
		UnchangedPasswordIntervalTicks: 200000000,
	}

	wire := MarshalManagedPasswordBlob(blob)

	// PreviousPasswordOffset should be 0.
	if off := binary.LittleEndian.Uint16(wire[10:12]); off != 0 {
		t.Fatalf("PreviousPasswordOffset = %d, want 0", off)
	}

	parsed, err := ParseManagedPasswordBlob(wire)
	if err != nil {
		t.Fatalf("ParseManagedPasswordBlob: %v", err)
	}
	if parsed.PreviousPassword != "" {
		t.Errorf("PreviousPassword = %q, want empty", parsed.PreviousPassword)
	}
	if parsed.CurrentPassword != "OnlyCurrentPwd" {
		t.Errorf("CurrentPassword = %q, want %q", parsed.CurrentPassword, "OnlyCurrentPwd")
	}
}

func TestManagedPasswordBlobTooShort(t *testing.T) {
	_, err := ParseManagedPasswordBlob(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short blob")
	}
}

func TestManagedPasswordBlobBadVersion(t *testing.T) {
	data := make([]byte, 32)
	binary.LittleEndian.PutUint16(data[0:2], 2) // bad version
	binary.LittleEndian.PutUint32(data[4:8], 32)
	_, err := ParseManagedPasswordBlob(data)
	if err == nil {
		t.Error("expected error for bad version")
	}
}

func TestManagedPasswordBlobLengthExceedsData(t *testing.T) {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	binary.LittleEndian.PutUint32(data[4:8], 1000) // length > data size
	_, err := ParseManagedPasswordBlob(data)
	if err == nil {
		t.Error("expected error for length exceeding data")
	}
}

func TestManagedPasswordBlobIntervalAlignment(t *testing.T) {
	blob := &ManagedPasswordBlob{
		CurrentPassword:                "A",
		QueryPasswordIntervalTicks:     12345,
		UnchangedPasswordIntervalTicks: 67890,
	}

	wire := MarshalManagedPasswordBlob(blob)

	queryOff := int(binary.LittleEndian.Uint16(wire[12:14]))
	unchangedOff := int(binary.LittleEndian.Uint16(wire[14:16]))

	if queryOff%8 != 0 {
		t.Errorf("QueryPasswordInterval offset %d is not 8-byte aligned", queryOff)
	}
	if unchangedOff%8 != 0 {
		t.Errorf("UnchangedPasswordInterval offset %d is not 8-byte aligned", unchangedOff)
	}

	parsed, err := ParseManagedPasswordBlob(wire)
	if err != nil {
		t.Fatalf("ParseManagedPasswordBlob: %v", err)
	}
	if parsed.QueryPasswordIntervalTicks != 12345 {
		t.Errorf("QueryPasswordIntervalTicks = %d, want 12345", parsed.QueryPasswordIntervalTicks)
	}
	if parsed.UnchangedPasswordIntervalTicks != 67890 {
		t.Errorf("UnchangedPasswordIntervalTicks = %d, want 67890", parsed.UnchangedPasswordIntervalTicks)
	}
}

func TestUTF16NullTermRoundTrip(t *testing.T) {
	for _, s := range []string{"", "hello", "P@ssw0rd!", "日本語テスト"} {
		enc := encodeUTF16NullTerm(s)
		dec, err := readUTF16NullTerm(enc, 0)
		if err != nil {
			t.Errorf("readUTF16NullTerm(%q): %v", s, err)
			continue
		}
		if dec != s {
			t.Errorf("round-trip %q -> %q", s, dec)
		}
	}
}
