package nbns

import (
	"testing"
)

// TestFirstLevelEncodeRFC1001KAT is the canonical known-answer test from
// RFC 1001 section 4.1: the 16-byte NetBIOS name "FRED" padded to 16 bytes
// with spaces (0x20) first-level encodes to "EGFCEFEECACACACACACACACACACACACA".
// Each nibble is split high-then-low and offset by ASCII 'A' (0x41); a space
// byte (0x20) therefore encodes to the pair "CA".
func TestFirstLevelEncodeRFC1001KAT(t *testing.T) {
	const want = "EGFCEFEECACACACACACACACACACACACA"

	n := &NetBIOSName{Name: "FRED"}
	got, err := n.FirstLevelEncode()
	if err != nil {
		t.Fatalf("FirstLevelEncode returned error: %v", err)
	}
	if got != want {
		t.Fatalf("FirstLevelEncode(%q) = %q, want %q", n.Name, got, want)
	}
	if len(got) != EncodedNameLength {
		t.Fatalf("encoded length = %d, want %d", len(got), EncodedNameLength)
	}
}

// TestFirstLevelDecodeRFC1001KAT is the inverse of the RFC 1001 4.1 example:
// decoding "EGFCEFEECACACACACACACACACACACACA" yields "FRED" (trailing space
// padding trimmed).
func TestFirstLevelDecodeRFC1001KAT(t *testing.T) {
	const encoded = "EGFCEFEECACACACACACACACACACACACA"

	nb, err := FirstLevelDecode(encoded)
	if err != nil {
		t.Fatalf("FirstLevelDecode returned error: %v", err)
	}
	if nb.Name != "FRED" {
		t.Fatalf("FirstLevelDecode(%q).Name = %q, want %q", encoded, nb.Name, "FRED")
	}
	if nb.ScopeID != "" {
		t.Fatalf("FirstLevelDecode(%q).ScopeID = %q, want empty", encoded, nb.ScopeID)
	}
}

// TestFirstLevelEncodeWithScopeID checks that a scope identifier is appended
// as an RFC 1001 4.1 domain suffix ("<encoded>.<scope>").
func TestFirstLevelEncodeWithScopeID(t *testing.T) {
	const want = "EGFCEFEECACACACACACACACACACACACA.NETBIOS.COM"

	n := &NetBIOSName{Name: "FRED", ScopeID: "NETBIOS.COM"}
	got, err := n.FirstLevelEncode()
	if err != nil {
		t.Fatalf("FirstLevelEncode returned error: %v", err)
	}
	if got != want {
		t.Fatalf("FirstLevelEncode with scope = %q, want %q", got, want)
	}

	nb, err := FirstLevelDecode(got)
	if err != nil {
		t.Fatalf("FirstLevelDecode returned error: %v", err)
	}
	if nb.Name != "FRED" {
		t.Fatalf("round-trip Name = %q, want %q", nb.Name, "FRED")
	}
	if nb.ScopeID != "NETBIOS.COM" {
		t.Fatalf("round-trip ScopeID = %q, want %q", nb.ScopeID, "NETBIOS.COM")
	}
}

// TestFirstLevelEncodeSuffixByte verifies that a name whose 16th byte carries
// a NetBIOS suffix (service identifier) survives encode/decode. Here the name
// occupies 15 bytes and the 16th byte is the workstation-service suffix 0x00.
// The trailing 0x00 encodes to "AA" and, unlike a space, is NOT trimmed on
// decode, so the decoded name retains the embedded suffix byte.
func TestFirstLevelEncodeSuffixByte(t *testing.T) {
	// 15-char base + a single explicit suffix byte in position 16.
	base := "WORKSTATION1234"
	name := base + string([]byte{0x00})

	n := &NetBIOSName{Name: name}
	encoded, err := n.FirstLevelEncode()
	if err != nil {
		t.Fatalf("FirstLevelEncode returned error: %v", err)
	}
	if len(encoded) != EncodedNameLength {
		t.Fatalf("encoded length = %d, want %d", len(encoded), EncodedNameLength)
	}
	// 0x00 -> high nibble 0 ('A'), low nibble 0 ('A').
	if last := encoded[len(encoded)-2:]; last != "AA" {
		t.Fatalf("suffix byte encoding = %q, want %q", last, "AA")
	}

	nb, err := FirstLevelDecode(encoded)
	if err != nil {
		t.Fatalf("FirstLevelDecode returned error: %v", err)
	}
	if nb.Name != name {
		t.Fatalf("round-trip Name = %q (% x), want %q (% x)", nb.Name, nb.Name, name, name)
	}
}

// TestFirstLevelRoundTripFuzz encodes then decodes a handful of names and
// asserts the space-trimmed value is recovered. Names shorter than 16 bytes
// are padded with spaces on encode and trimmed on decode, so the recovered
// value equals the original with trailing spaces removed.
func TestFirstLevelRoundTripFuzz(t *testing.T) {
	names := []string{
		"A",
		"FRED",
		"WORKSTATION",
		"SIXTEENCHARLONG!", // exactly 16 bytes
		"HOST-01",
		"*SMBSERVER",
	}

	for _, name := range names {
		// "*SMBSERVER" starts with '*', which Validate rejects; skip it as
		// an encode input but keep it documented as an invalid case below.
		if name == "*SMBSERVER" {
			n := &NetBIOSName{Name: name}
			if _, err := n.FirstLevelEncode(); err == nil {
				t.Errorf("FirstLevelEncode(%q) expected error (name cannot start with *)", name)
			}
			continue
		}

		n := &NetBIOSName{Name: name}
		encoded, err := n.FirstLevelEncode()
		if err != nil {
			t.Errorf("FirstLevelEncode(%q) error: %v", name, err)
			continue
		}

		nb, err := FirstLevelDecode(encoded)
		if err != nil {
			t.Errorf("FirstLevelDecode(%q) error: %v", encoded, err)
			continue
		}

		if nb.Name != name {
			t.Errorf("round-trip %q -> %q -> %q", name, encoded, nb.Name)
		}
	}
}

// TestFirstLevelEncodeTooLong ensures a name exceeding 16 bytes is rejected.
func TestFirstLevelEncodeTooLong(t *testing.T) {
	n := &NetBIOSName{Name: "THISNAMEISWAYTOOLONG"}
	if _, err := n.FirstLevelEncode(); err == nil {
		t.Fatalf("FirstLevelEncode expected error for over-long name")
	}
}

// TestFirstLevelDecodeErrors covers malformed encoded inputs: wrong length and
// characters outside the RFC 1001 'A'..'P' half-ASCII alphabet.
func TestFirstLevelDecodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{"too short", "EGFC"},
		{"too long", "EGFCEFEECACACACACACACACACACACACAA"},
		{"empty", ""},
		{"out of range char", "EGFCEFEECACACACACACACACACACACAC0"}, // '0' < 'A'
		{"high char", "ZGFCEFEECACACACACACACACACACACACA"},         // 'Z'-'A' = 25 > 0x0F
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := FirstLevelDecode(tc.encoded); err == nil {
				t.Fatalf("FirstLevelDecode(%q) expected error, got nil", tc.encoded)
			}
		})
	}
}
