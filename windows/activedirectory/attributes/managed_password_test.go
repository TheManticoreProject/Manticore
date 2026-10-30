package attributes

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/md4"
)

func TestManagedPasswordBlobRoundTrip(t *testing.T) {
	blob := &ManagedPasswordBlob{
		CurrentPassword:                "P@ssw0rd!",
		PreviousPassword:               "OldP@ss",
		QueryPasswordIntervalTicks:     864000000000,  // 1 day in 100ns ticks
		UnchangedPasswordIntervalTicks: 2592000000000, // 30 days
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

func TestManagedPasswordBlobIntervalLayout(t *testing.T) {
	blob := &ManagedPasswordBlob{
		CurrentPassword:                "A",
		QueryPasswordIntervalTicks:     12345,
		UnchangedPasswordIntervalTicks: 67890,
	}

	wire := MarshalManagedPasswordBlob(blob)

	queryOff := int(binary.LittleEndian.Uint16(wire[12:14]))
	unchangedOff := int(binary.LittleEndian.Uint16(wire[14:16]))

	// DCs place the intervals directly after the last password terminator.
	if want := managedPasswordHeaderSize + 4; queryOff != want {
		t.Errorf("QueryPasswordInterval offset = %d, want %d", queryOff, want)
	}
	if unchangedOff != queryOff+8 {
		t.Errorf("UnchangedPasswordInterval offset = %d, want %d", unchangedOff, queryOff+8)
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

// gmsaPasswordBytes returns 256 bytes of gMSA-like password material containing
// an unpaired high surrogate (0xD801) followed by 'A', and an embedded 0x0000
// code unit, neither of which survives a UTF-16 string decode. When
// trailingZeros is set, the last two code units are 0x0000.
func gmsaPasswordBytes(seed byte, trailingZeros bool) []byte {
	pw := make([]byte, 256)
	for i := range pw {
		pw[i] = byte(i)*7 + seed + 1
	}
	copy(pw[10:], []byte{0x01, 0xd8, 0x41, 0x00}) // unpaired surrogate, then 'A'
	copy(pw[40:], []byte{0x00, 0x00})             // embedded NUL code unit
	if trailingZeros {
		copy(pw[252:], []byte{0x00, 0x00, 0x00, 0x00})
	}
	return pw
}

// buildGMSABlob lays out an MSDS-MANAGEDPASSWORD_BLOB the way a DC does,
// without alignment padding:
// header | current+NUL | [previous+NUL] | query | unchanged.
func buildGMSABlob(current, previous []byte) []byte {
	currentOff := managedPasswordHeaderSize
	previousOff := currentOff + len(current) + 2
	queryOff := previousOff
	if previous != nil {
		queryOff = previousOff + len(previous) + 2
	}
	total := queryOff + 16
	buf := make([]byte, total)
	binary.LittleEndian.PutUint16(buf[0:2], 1)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(total))
	binary.LittleEndian.PutUint16(buf[8:10], uint16(currentOff))
	if previous != nil {
		binary.LittleEndian.PutUint16(buf[10:12], uint16(previousOff))
		copy(buf[previousOff:], previous)
	}
	binary.LittleEndian.PutUint16(buf[12:14], uint16(queryOff))
	binary.LittleEndian.PutUint16(buf[14:16], uint16(queryOff+8))
	copy(buf[currentOff:], current)
	binary.LittleEndian.PutUint64(buf[queryOff:], 111)
	binary.LittleEndian.PutUint64(buf[queryOff+8:], 222)
	return buf
}

func TestManagedPasswordBlobRawPasswordBytes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		current  []byte
		previous []byte
	}{
		{"current only", gmsaPasswordBytes(0, false), nil},
		{"current and previous", gmsaPasswordBytes(0, false), gmsaPasswordBytes(0x55, false)},
		{"current only, trailing zero units", gmsaPasswordBytes(0, true), nil},
		{"current and previous, trailing zero units", gmsaPasswordBytes(0, true), gmsaPasswordBytes(0x55, true)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := buildGMSABlob(tc.current, tc.previous)

			parsed, err := ParseManagedPasswordBlob(wire)
			if err != nil {
				t.Fatalf("ParseManagedPasswordBlob: %v", err)
			}
			if !bytes.Equal(parsed.CurrentPasswordBytes, tc.current) {
				t.Errorf("CurrentPasswordBytes = %x, want %x", parsed.CurrentPasswordBytes, tc.current)
			}
			if !bytes.Equal(parsed.PreviousPasswordBytes, tc.previous) {
				t.Errorf("PreviousPasswordBytes = %x, want %x", parsed.PreviousPasswordBytes, tc.previous)
			}

			// The NT hash is MD4 over the raw UTF-16LE password bytes.
			if got, want := md4.Sum(parsed.CurrentPasswordBytes), md4.Sum(tc.current); got != want {
				t.Errorf("current NT hash = %x, want %x", got, want)
			}

			// Re-encoding must reproduce the original wire bytes exactly.
			if again := MarshalManagedPasswordBlob(parsed); !bytes.Equal(again, wire) {
				t.Errorf("re-marshalled blob differs from original:\n got %x\nwant %x", again, wire)
			}
		})
	}
}

// TestManagedPasswordBlobDCVector uses a real DC-produced msDS-ManagedPassword
// value and its NT hash from the DSInternals test suite
// (Src/DSInternals.Common.Test/ManagedPasswordTester.cs).
func TestManagedPasswordBlobDCVector(t *testing.T) {
	wire, err := hex.DecodeString("01000000220100001000000012011a011609f270f541c315ffee9fcd22a98447b5c6e6fb7151cb020a2b017bb4e003647949967fc96f7c9ec3426b80901bb9c162867cbc68c520c4d7a431c3d9a670f8aa41d2ae5c0c08f27f8698b90c18a5a576e9933fb0cadaf8e661be2f58308c580866b1ae582ee50a9aa7c5d65a312dbbc3542c51c7e0b2d4c61e9763de481d9963367273aa72b53c2e402e31c6cd38e7785ad06639cdfa07738d19ae20c370e06787ad2f600823c505fc9dd32b3f06505da37b86b298d3650140af83c1f01c907964d182ea0efb19e74c949f58123fdecb41f78ed0eabbde31bb46afd3134da82550380ed36038d100f71095404a97e52d661dbe4f74deef4122a102dca698960000864973f96017000086eba24660170000")
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseManagedPasswordBlob(wire)
	if err != nil {
		t.Fatalf("ParseManagedPasswordBlob: %v", err)
	}
	if len(parsed.CurrentPasswordBytes) != 256 {
		t.Errorf("len(CurrentPasswordBytes) = %d, want 256", len(parsed.CurrentPasswordBytes))
	}
	if parsed.PreviousPasswordBytes != nil {
		t.Errorf("PreviousPasswordBytes = %x, want nil", parsed.PreviousPasswordBytes)
	}
	ntHash := md4.Sum(parsed.CurrentPasswordBytes)
	if got, want := hex.EncodeToString(ntHash[:]), "1fe07f47bfa7f511d902ed5cfb79cc4d"; got != want {
		t.Errorf("current NT hash = %s, want %s", got, want)
	}
	if again := MarshalManagedPasswordBlob(parsed); !bytes.Equal(again, wire) {
		t.Errorf("re-marshalled blob differs from original:\n got %x\nwant %x", again, wire)
	}
}
