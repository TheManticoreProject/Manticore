package auth

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Known-answer vectors from iSeriesNavigatorReverse
// .research/family-bf/signon-crypto-vectors.txt (vector-file SHA-256
// f92a7f65ae22113d255b2de62af70a482eb79c5c5182cb30ae81c3e2a7268db3),
// produced by the installed JTOpen 4.8 jar as a black-box oracle. All
// profiles, passwords and seeds are synthetic.
var (
	kaProfile    = "DEMOÑUSR"
	kaClientSeed = [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	kaServerSeed = [8]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
)

func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSignonTableCanonical(t *testing.T) {
	in := "#$0123456789@ABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	want := unhex(t, "7b5bf0f1f2f3f4f5f6f7f8f97cc1c2c3c4c5c6c7c8c9d1d2d3d4d5d6d7d8d9e2e3e4e5e6e7e8e96d")
	for i, r := range in {
		b, c, ok := mapSignonRune(r)
		if !ok || b != want[i] || c != r {
			t.Errorf("map(%q) = %02x, %q, %t; want %02x", r, b, c, ok, want[i])
		}
	}
}

func TestSignonTableAliases(t *testing.T) {
	aliases := []rune{0x00a3, 0x00a5, 0x00a7, 0x00c4, 0x00c5, 0x00c6, 0x00d0, 0x00d1, 0x00d6, 0x00d8, 0x00e0, 0x0130, 0x015e}
	wantBytes := unhex(t, "7b5b7c7b5b7b7c7b7c7c7c5b7c")
	wantCanonical := []rune{'#', '$', '@', '#', '$', '#', '@', '#', '@', '@', '@', '$', '@'}
	for i, r := range aliases {
		b, c, ok := mapSignonRune(r)
		if !ok || b != wantBytes[i] || c != wantCanonical[i] {
			t.Errorf("alias %U -> %02x %q %t; want %02x %q", r, b, c, ok, wantBytes[i], wantCanonical[i])
		}
	}
	for _, r := range []rune{' ', '?', 'a', 'é', 0} {
		if _, _, ok := mapSignonRune(r); ok {
			t.Errorf("%U accepted; the table is closed", r)
		}
	}
}

func TestEncodeProfile(t *testing.T) {
	p, err := EncodeProfile(kaProfile)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(p.EBCDIC[:]); got != "c4c5d4d67be4e2d94040" {
		t.Errorf("EBCDIC = %s", got)
	}
	if got := hex.EncodeToString(p.UTF16BE()); got != "00440045004d004f002300550053005200200020" {
		t.Errorf("canonical UTF-16BE = %s", got)
	}
	lower, err := EncodeProfile("demoñusr")
	if err != nil || lower.EBCDIC != p.EBCDIC {
		t.Errorf("lowercase profile = %x, %v", lower.EBCDIC, err)
	}
	for _, bad := range []string{"", "ABCDEFGHIJK", "A B", "USER?", "USÉR"} {
		_, err := EncodeProfile(bad)
		if !errors.Is(err, ErrInvalidProfile) {
			t.Errorf("EncodeProfile(%q) error = %v", bad, err)
		}
		if bad != "" && strings.Contains(err.Error(), bad) {
			t.Errorf("error leaks the profile: %v", err)
		}
	}
	if _, err := EncodeProfile("ABCDEFGHIJ"); err != nil {
		t.Errorf("10-character profile rejected: %v", err)
	}
}

func TestDESKnownAnswer(t *testing.T) {
	p, _ := EncodeProfile(kaProfile)
	pw, err := encodeDESPassword([]byte("SYNTHD3S"))
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(pw[:]); got != "e2e8d5e3c8c4f3e24040" {
		t.Fatalf("DES password encoding = %s", got)
	}
	token := desPasswordToken(p.EBCDIC, pw)
	if got := hex.EncodeToString(token[:]); got != "5ec7804759c0a62d" {
		t.Fatalf("DES token = %s", got)
	}
	sub := desSubstitute(token, p.EBCDIC, kaClientSeed, kaServerSeed)
	if got := hex.EncodeToString(sub[:]); got != "52817d0e257aae4c" {
		t.Fatalf("DES substitute = %s", got)
	}
}

func TestSHA1KnownAnswer(t *testing.T) {
	p, _ := EncodeProfile(kaProfile)
	pw, err := shaPasswordUTF16BE([]byte("Synthetic42"))
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(pw); got != "00530079006e00740068006500740069006300340032" {
		t.Fatalf("SHA password encoding = %s", got)
	}
	token := sha1Token(p.UTF16BE(), pw)
	if got := hex.EncodeToString(token[:]); got != "8689d6f3bbe9726640d70f91404a96763b6fd0cf" {
		t.Fatalf("SHA-1 token = %s", got)
	}
	sub := sha1Substitute(token, p.UTF16BE(), kaClientSeed, kaServerSeed)
	if got := hex.EncodeToString(sub[:]); got != "114fbb27bb186485557b8b74725a24671e42ff8e" {
		t.Fatalf("SHA-1 substitute = %s", got)
	}
}

func TestDeriveMatchesKnownAnswers(t *testing.T) {
	for _, tc := range []struct {
		family   Family
		password string
		credType byte
		want     string
	}{
		{FamilyDES, "SYNTHD3S", CredentialTypeDES, "52817d0e257aae4c"},
		{FamilyDES, "synthd3s", CredentialTypeDES, "52817d0e257aae4c"}, // level 0/1 uppercases
		{FamilySHA1, "Synthetic42", CredentialTypeSHA1, "114fbb27bb186485557b8b74725a24671e42ff8e"},
		{FamilySHA1, "Synthetic42  　\x00", CredentialTypeSHA1, "114fbb27bb186485557b8b74725a24671e42ff8e"},
	} {
		c, err := NewPasswordCredential(kaProfile, []byte(tc.password))
		if err != nil {
			t.Fatal(err)
		}
		s, err := Derive(c, tc.family, kaClientSeed, kaServerSeed)
		if err != nil {
			t.Fatalf("%s %q: %v", tc.family, tc.password, err)
		}
		if s.CredentialType != tc.credType || hex.EncodeToString(s.Value) != tc.want || hex.EncodeToString(s.Profile[:]) != "c4c5d4d67be4e2d94040" {
			t.Errorf("%s %q: type %02x value %x", tc.family, tc.password, s.CredentialType, s.Value)
		}
	}
}

func TestDESPasswordRules(t *testing.T) {
	if pw, err := encodeDESPassword([]byte("123456789")); err != nil || hex.EncodeToString(pw[:]) != "d8f1f2f3f4f5f6f7f8f9" {
		t.Errorf("9-digit password = %x, %v; want Q-prefixed", pw, err)
	}
	if pw, err := encodeDESPassword([]byte("1ABC")); err != nil || pw[0] != 0xF1 {
		t.Errorf("leading digit only = %x, %v; JTOpen 4.8 prefixes only all-digit passwords", pw, err)
	}
	for name, pw := range map[string]string{
		"ten digits become eleven": "1234567890",
		"eleven characters":        "ABCDEFGHIJK",
		"empty":                    "",
		"space":                    "AB CD",
		"unsupported character":    "PASS?WORD",
		"invalid UTF-8":            "\xff\xfe",
	} {
		if _, err := encodeDESPassword([]byte(pw)); !errors.Is(err, ErrInvalidPassword) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	// A 10-character password exercises the second token.
	p, _ := EncodeProfile("ABCDEFGHIJ")
	ten, err := encodeDESPassword([]byte("PASSWORD12"))
	if err != nil {
		t.Fatal(err)
	}
	eight, _ := encodeDESPassword([]byte("PASSWORD"))
	if desPasswordToken(p.EBCDIC, ten) == desPasswordToken(p.EBCDIC, eight) {
		t.Error("characters 9-10 did not affect the token")
	}
}

func TestFoldProfile(t *testing.T) {
	short, _ := EncodeProfile("ABCDEFGH")
	if f := foldProfile(short.EBCDIC); !bytes.Equal(f[:], short.EBCDIC[:8]) {
		t.Errorf("8-character profile folded: %x", f)
	}
	long, _ := EncodeProfile("ABCDEFGHIJ") // I = C9, J = D1
	f := foldProfile(long.EBCDIC)
	want := []byte{
		0xC1 ^ (0xC9 & 0xC0), 0xC2 ^ (0xC9&0x30)<<2, 0xC3 ^ (0xC9&0x0C)<<4, 0xC4 ^ (0xC9&0x03)<<6,
		0xC5 ^ (0xD1 & 0xC0), 0xC6 ^ (0xD1&0x30)<<2, 0xC7 ^ (0xD1&0x0C)<<4, 0xC8 ^ (0xD1&0x03)<<6,
	}
	if !bytes.Equal(f[:], want) {
		t.Errorf("folded = %x, want %x", f, want)
	}
}

func TestSHAPasswordRules(t *testing.T) {
	for name, pw := range map[string]string{
		"leading star":  "*NONE",
		"empty":         "",
		"all blanks":    "   　",
		"invalid UTF-8": "\xff",
		"too long":      strings.Repeat("a", 129),
	} {
		if _, err := shaPasswordUTF16BE([]byte(pw)); !errors.Is(err, ErrInvalidPassword) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	if _, err := shaPasswordUTF16BE([]byte(strings.Repeat("a", 128))); err != nil {
		t.Errorf("128 characters rejected: %v", err)
	}
	if b, err := shaPasswordUTF16BE([]byte(" lead")); err != nil || len(b) != 10 {
		t.Errorf("leading space must be kept: %x, %v", b, err)
	}
}

func TestFamilyForLevel(t *testing.T) {
	for level, want := range map[byte]Family{0: FamilyDES, 1: FamilyDES, 2: FamilySHA1, 3: FamilySHA1} {
		if f, err := FamilyForLevel(level); err != nil || f != want {
			t.Errorf("level %d = %s, %v", level, f, err)
		}
	}
	for _, level := range []byte{4, 5, 255} {
		if _, err := FamilyForLevel(level); !errors.Is(err, ErrUnsupportedPasswordLevel) {
			t.Errorf("level %d accepted", level)
		}
	}
}

// fixedReader replays byte chunks, then zeros.
type fixedReader struct{ chunks [][]byte }

func (r *fixedReader) Read(p []byte) (int, error) {
	if len(r.chunks) == 0 {
		return copy(p, make([]byte, len(p))), nil
	}
	n := copy(p, r.chunks[0])
	r.chunks = r.chunks[1:]
	return n, nil
}

func TestNewSeedRange(t *testing.T) {
	zero := make([]byte, 8)
	tooBig := unhex(t, "e000000000000000")
	maxOK := unhex(t, "dfffffffffffffff")
	r := &fixedReader{chunks: [][]byte{zero, tooBig, maxOK}}
	seed, err := NewSeed(r)
	if err != nil || !bytes.Equal(seed[:], maxOK) {
		t.Fatalf("NewSeed = %x, %v; want the first in-range draw", seed, err)
	}
	if !ValidSeed([8]byte{0, 0, 0, 0, 0, 0, 0, 1}) || ValidSeed([8]byte{}) || ValidSeed([8]byte{0xFF}) {
		t.Error("ValidSeed boundaries wrong")
	}
	if _, err := NewSeed(&fixedReader{}); !errors.Is(err, ErrSeedSource) {
		t.Errorf("always-zero source error = %v", err)
	}
	for i := 0; i < 100; i++ {
		s, err := NewSeed(nil)
		if err != nil || !ValidSeed(s) {
			t.Fatalf("crypto/rand seed %x, %v", s, err)
		}
	}
}

func TestCredentialRedactionAndDestroy(t *testing.T) {
	const secret = "CANARYPW9"
	c, err := NewPasswordCredential("CANARYUSR", []byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X", "%d", "%T"} {
		out := fmt.Sprintf(verb, c)
		if verb != "%T" && out != redacted {
			t.Errorf("%s = %q", verb, out)
		}
		for _, leak := range []string{secret, "CANARY", "c3c1d5c1d9e8", "43414e415259"} {
			if strings.Contains(strings.ToLower(out), strings.ToLower(leak)) {
				t.Errorf("%s leaks %q: %s", verb, leak, out)
			}
		}
	}
	s, _ := Derive(c, FamilyDES, kaClientSeed, kaServerSeed)
	for _, out := range []string{fmt.Sprint(s), fmt.Sprintf("%#v", s), fmt.Sprintf("%+v", s)} {
		if strings.Contains(out, hex.EncodeToString(s.Value)) || !strings.Contains(out, "redacted") {
			t.Errorf("substitute formatting leaks: %s", out)
		}
	}
	s.Wipe()
	if !bytes.Equal(s.Value, make([]byte, 8)) {
		t.Error("Wipe left substitute bytes")
	}

	input := []byte(secret)
	c2, _ := NewPasswordCredential("USER", input)
	input[0] = 'X'
	s2, _ := Derive(c2, FamilyDES, kaClientSeed, kaServerSeed)
	c3, _ := NewPasswordCredential("USER", []byte(secret))
	s3, _ := Derive(c3, FamilyDES, kaClientSeed, kaServerSeed)
	if !bytes.Equal(s2.Value, s3.Value) {
		t.Error("the credential did not copy its password input")
	}

	c.Destroy()
	c.Destroy()
	if _, err := Derive(c, FamilyDES, kaClientSeed, kaServerSeed); !errors.Is(err, ErrCredentialDestroyed) {
		t.Errorf("Derive after Destroy = %v", err)
	}
	if c.password != nil || c.profile.EBCDIC != ([10]byte{}) {
		t.Error("Destroy left secret material")
	}
}

func TestNewPasswordCredentialRejects(t *testing.T) {
	if _, err := NewPasswordCredential("BAD USER", []byte("x")); !errors.Is(err, ErrInvalidProfile) {
		t.Errorf("bad profile error = %v", err)
	}
	if _, err := NewPasswordCredential("USER", nil); !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("empty password error = %v", err)
	}
	c, _ := NewPasswordCredential("USER", []byte("1234567890"))
	if _, err := Derive(c, FamilyDES, kaClientSeed, kaServerSeed); !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("10-digit DES password error = %v", err)
	}
	if _, err := Derive(c, Family(9), kaClientSeed, kaServerSeed); err == nil {
		t.Error("unknown family accepted")
	}
}
