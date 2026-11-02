package lm

import (
	"encoding/hex"
	"errors"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

// windowsLMVectors were computed on Windows Server 2012 R2 (OEM code page 437)
// with RtlUpcaseUnicodeStringToOemString followed by SystemFunction006, the LM
// one-way function, so they are Windows' own conversion and hash.
var windowsLMVectors = []struct{ password, oem, hash string }{
	{"Passw0rd!", "504153535730524421", "b34ce522c3e4c87722c34254e51bff62"},
	{"password", "50415353574f5244", "e52cac67419a9a224a3b108f3fa6cb6d"},
	{"Pässwörd1", "508e53535799524431", "6b396da2d20f20b338f10713b629b565"},
	// È, À and Ù are not in CP437; Windows best-fits them to E, A, U.
	{"éèàçù", "9045418055", "b4cf770f57f57bfbaad3b435b51404ee"},
	{"ABCDEFGHIJKLMNOP", "4142434445464748494a4b4c4d4e4f50", "e0c510199cc66abd8c51ec214bebdea1"},
	{"abcdefghijklméx", "4142434445464748494a4b4c4d9058", "e0c510199cc66abd330ab407f15b3058"},
}

// TestLMHashMatchesWindows checks the OEM conversion and the LM hash of each
// password against the values Windows computed. Non-ASCII passwords used to be
// hashed from their UTF-8 bytes.
func TestLMHashMatchesWindows(t *testing.T) {
	for _, v := range windowsLMVectors {
		oem, err := toUpperOEM(v.password, charmap.CodePage437)
		if err != nil {
			t.Fatalf("%q: %v", v.password, err)
		}
		if got := hex.EncodeToString(oem); got != v.oem {
			t.Errorf("%q: OEM bytes %s, Windows %s", v.password, got, v.oem)
		}
		h := LMHash(v.password)
		if got := hex.EncodeToString(h[:]); got != v.hash {
			t.Errorf("%q: LM hash %s, Windows %s", v.password, got, v.hash)
		}
	}
}

// TestLMHashUnmappableCharacter checks a password Windows cannot convert to the
// OEM code page ("naïve€": € is not in CP437, and Windows' conversion fails with
// STATUS_UNMAPPABLE_CHARACTER) reports the error from LMHashOEM and has the
// empty LM hash from LMHash.
func TestLMHashUnmappableCharacter(t *testing.T) {
	if _, err := LMHashOEM("naïve€", charmap.CodePage437); !errors.Is(err, ErrUnmappableCharacter) {
		t.Errorf("LMHashOEM error = %v, want ErrUnmappableCharacter", err)
	}
	if h := LMHash("naïve€"); hex.EncodeToString(h[:]) != "aad3b435b51404eeaad3b435b51404ee" {
		t.Errorf("LMHash = %x, want the empty LM hash", h)
	}
	// Code page 1252 does have €.
	if _, err := LMHashOEM("naïve€", charmap.Windows1252); err != nil {
		t.Errorf("LMHashOEM(cp1252) = %v", err)
	}
}

// windowsOEMChars are single characters converted by Windows' own
// RtlUpcaseUnicodeStringToOemString on code page 437. "" means Windows refused
// the character (STATUS_UNMAPPABLE_CHARACTER).
var windowsOEMChars = []struct {
	r   rune
	oem string
}{
	{'µ', "e6"}, // not upper-cased to Greek Μ
	{'ı', "49"}, // best fit to I
	{'ſ', ""},   // not upper-cased to S
	{'ÿ', "59"}, // upper-cased to Ÿ, best fit to Y
	{'Ø', "4f"}, // best fit to O
	{'ø', "4f"}, // upper-cased to Ø, best fit to O
	{'ß', "e1"}, // kept as ß
	{'ς', ""},   // not upper-cased to Σ
	{'Ǆ', ""},   // no best fit
	{'ﬁ', ""},   // no best fit
	{'ŉ', ""},   // no best fit
}

// TestToUpperOEMMatchesWindowsCharacters checks the case mapping and best fit
// of characters where Unicode's upper-casing and Windows' differ.
func TestToUpperOEMMatchesWindowsCharacters(t *testing.T) {
	for _, c := range windowsOEMChars {
		oem, err := toUpperOEM(string(c.r), charmap.CodePage437)
		got := hex.EncodeToString(oem)
		if err != nil {
			got = ""
		}
		if got != c.oem {
			t.Errorf("%q (U+%04X): OEM %q, Windows %q (err %v)", c.r, c.r, got, c.oem, err)
		}
	}
}
