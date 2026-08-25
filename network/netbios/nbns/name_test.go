package nbns

import (
	"bytes"
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

// TestEncodeSessionServiceName verifies the 34-byte second-level encoding used
// by the NetBIOS session service, including the "*SMBSERVER" wildcard convention
// and the workstation/server service suffixes.
func TestEncodeSessionServiceName(t *testing.T) {
	// "*SMBSERVER" with the server service suffix 0x20: 0x20 length prefix, the
	// well-known 32-char first-level encoding, and a 0x00 terminator.
	got, err := EncodeSessionServiceName("*SMBSERVER", 0x20)
	if err != nil {
		t.Fatalf("EncodeSessionServiceName() error = %v", err)
	}
	want := append(append([]byte{0x20}, []byte("CKFDENECFDEFFCFGEFFCCACACACACACA")...), 0x00)
	if !bytes.Equal(got, want) {
		t.Fatalf("encoded *SMBSERVER = % x, want % x", got, want)
	}
	if len(got) != 34 {
		t.Fatalf("encoded length = %d, want 34", len(got))
	}

	// A workstation name (suffix 0x00) must place the suffix byte in the final
	// (16th) name position, which decodes to a trailing 0x00.
	wks, err := EncodeSessionServiceName("CLIENT", 0x00)
	if err != nil {
		t.Fatalf("EncodeSessionServiceName() error = %v", err)
	}
	if wks[0] != 0x20 || wks[33] != 0x00 {
		t.Fatalf("workstation framing = % x, want 0x20 ... 0x00", wks)
	}
	// The final encoded byte pair (positions 31-32) encodes the suffix 0x00 as
	// two 'A' characters.
	if wks[31] != 'A' || wks[32] != 'A' {
		t.Fatalf("suffix encoding = %q%q, want AA", wks[31], wks[32])
	}

	// Names longer than 15 characters are rejected.
	if _, err := EncodeSessionServiceName("THISNAMEISWAYTOOLONG", 0x00); err == nil {
		t.Fatal("EncodeSessionServiceName() should reject names longer than 15 characters")
	}
}

// TestDecodeSessionServiceName asserts the exact wire decoding of a
// second-level-encoded session-service name, including the RFC 1001 known
// answer, the suffix split and the byte count consumed.
func TestDecodeSessionServiceName(t *testing.T) {
	// The "*SMBSERVER" / 0x20 encoding from TestEncodeSessionServiceName.
	encoded := append(append([]byte{0x20}, []byte("CKFDENECFDEFFCFGEFFCCACACACACACA")...), 0x00)

	name, suffix, n, err := DecodeSessionServiceName(encoded)
	if err != nil {
		t.Fatalf("DecodeSessionServiceName() error = %v", err)
	}
	if name != "*SMBSERVER" {
		t.Fatalf("name = %q, want %q", name, "*SMBSERVER")
	}
	if suffix != 0x20 {
		t.Fatalf("suffix = 0x%02X, want 0x20", suffix)
	}
	if n != 34 {
		t.Fatalf("consumed = %d, want 34", n)
	}

	// Trailing bytes must be left untouched: a SESSION REQUEST carries the
	// CALLING name immediately after the CALLED name, so the caller decodes the
	// second name from encoded[n:].
	twoNames := append(append([]byte{}, encoded...), encoded...)
	_, _, first, err := DecodeSessionServiceName(twoNames)
	if err != nil {
		t.Fatalf("DecodeSessionServiceName() on a two-name buffer error = %v", err)
	}
	second, secondSuffix, _, err := DecodeSessionServiceName(twoNames[first:])
	if err != nil {
		t.Fatalf("DecodeSessionServiceName() on the second name error = %v", err)
	}
	if second != "*SMBSERVER" || secondSuffix != 0x20 {
		t.Fatalf("second name = %q/0x%02X, want *SMBSERVER/0x20", second, secondSuffix)
	}
}

// TestDecodeSessionServiceNameRoundTrip asserts DecodeSessionServiceName is the
// exact inverse of EncodeSessionServiceName across name lengths and suffixes,
// including the space padding the encoder applies.
func TestDecodeSessionServiceNameRoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		suffix byte
	}{
		{"A", 0x00},
		{"CLIENT", 0x00},
		{"FILESERVER01", 0x20},
		{"*SMBSERVER", 0x20},
		{"WORKSTATION1234", 0x03}, // exactly 15 characters, the maximum
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeSessionServiceName(tc.name, tc.suffix)
			if err != nil {
				t.Fatalf("EncodeSessionServiceName() error = %v", err)
			}
			got, suffix, n, err := DecodeSessionServiceName(encoded)
			if err != nil {
				t.Fatalf("DecodeSessionServiceName() error = %v", err)
			}
			if got != tc.name {
				t.Fatalf("name = %q, want %q", got, tc.name)
			}
			if suffix != tc.suffix {
				t.Fatalf("suffix = 0x%02X, want 0x%02X", suffix, tc.suffix)
			}
			if n != len(encoded) {
				t.Fatalf("consumed = %d, want %d", n, len(encoded))
			}
		})
	}
}

// TestDecodeSessionServiceNameScoped asserts a name carrying scope labels is
// decoded and that the consumed count covers the whole label sequence, so a
// following name is still found at the right offset.
func TestDecodeSessionServiceNameScoped(t *testing.T) {
	base, err := EncodeSessionServiceName("SERVER", 0x20)
	if err != nil {
		t.Fatalf("EncodeSessionServiceName() error = %v", err)
	}
	// Replace the bare 0x00 terminator with two scope labels ("EXAMPLE.COM")
	// followed by the terminator.
	scoped := append([]byte{}, base[:len(base)-1]...)
	scoped = append(scoped, 7)
	scoped = append(scoped, []byte("EXAMPLE")...)
	scoped = append(scoped, 3)
	scoped = append(scoped, []byte("COM")...)
	scoped = append(scoped, 0x00)

	name, suffix, n, err := DecodeSessionServiceName(scoped)
	if err != nil {
		t.Fatalf("DecodeSessionServiceName() on a scoped name error = %v", err)
	}
	if name != "SERVER" || suffix != 0x20 {
		t.Fatalf("name = %q/0x%02X, want SERVER/0x20", name, suffix)
	}
	if n != len(scoped) {
		t.Fatalf("consumed = %d, want %d", n, len(scoped))
	}
}

// TestDecodeSessionServiceNameErrors asserts each malformed encoding is rejected
// rather than producing a garbage name or reading out of bounds.
func TestDecodeSessionServiceNameErrors(t *testing.T) {
	valid, err := EncodeSessionServiceName("SERVER", 0x20)
	if err != nil {
		t.Fatalf("EncodeSessionServiceName() error = %v", err)
	}

	// A length byte other than 0x20.
	badLength := append([]byte{}, valid...)
	badLength[0] = 0x10

	// An encoding character outside the 'A'..'P' half-byte alphabet.
	badChar := append([]byte{}, valid...)
	badChar[1] = 'Z'

	// No label terminator at all.
	unterminated := append([]byte{}, valid[:len(valid)-1]...)

	// A scope label whose length runs past the end of the buffer.
	truncatedLabel := append(append([]byte{}, valid[:len(valid)-1]...), 40, 'A', 'B')

	// A scope label longer than the 63-byte DNS limit.
	oversizeLabel := append(append([]byte{}, valid[:len(valid)-1]...), 64)

	cases := []struct {
		name  string
		input []byte
	}{
		{"empty", nil},
		{"truncated", valid[:20]},
		{"bad length byte", badLength},
		{"bad encoding character", badChar},
		{"unterminated", unterminated},
		{"truncated scope label", truncatedLabel},
		{"oversize scope label", oversizeLabel},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := DecodeSessionServiceName(tc.input); err == nil {
				t.Fatalf("DecodeSessionServiceName(% x) should fail", tc.input)
			}
		})
	}
}
