package ebcdic

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/encoding/ebcdic/cp500"
)

// decodeTable returns the rune each byte of a codec decodes to.
func decodeTable(t *testing.T, cp int) [256]rune {
	t.Helper()
	c, ok := IBMCodePage(cp)
	if !ok {
		t.Fatalf("IBMCodePage(%d) not found", cp)
	}
	var table [256]rune
	for b := range table {
		s, _ := c.Decode([]byte{byte(b)})
		table[b] = []rune(s)[0]
	}
	return table
}

// TestEuroVariantsDifferOnlyAtTheEuro checks each 114x code page against the
// national code page it extends: IBM replaced one character with the euro sign
// and changed nothing else. The tables come from separate ICU files, so this
// also catches a stale or mismatched source.
func TestEuroVariantsDifferOnlyAtTheEuro(t *testing.T) {
	pairs := map[int]int{1140: 37, 1141: 273, 1142: 277, 1143: 278, 1144: 280, 1145: 284, 1146: 285, 1147: 297, 1148: 500, 1149: 871}
	for euro, base := range pairs {
		e, b := decodeTable(t, euro), decodeTable(t, base)
		var diffs []int
		for i := range e {
			if e[i] != b[i] {
				diffs = append(diffs, i)
			}
		}
		if len(diffs) != 1 || e[diffs[0]] != '€' || b[diffs[0]] != '¤' {
			t.Errorf("CCSID %d differs from %d at bytes %x, want only the euro replacing ¤", euro, base, diffs)
		}
	}
}

// TestGeneratedTablesAgreeWithOtherSources compares the generated 500 table
// with the package's legacy cp500 byte map, an independent copy of the same
// code page, for every byte that map carries as Latin-1.
func TestGeneratedTablesAgreeWithOtherSources(t *testing.T) {
	t500 := decodeTable(t, 500)
	for b, latin1 := range cp500.EbcdicCp500ToAsciiMap {
		if t500[b] < 0x100 && t500[b] != rune(latin1) {
			t.Errorf("500 byte 0x%02X = %U, legacy map %#02x", b, t500[b], latin1)
		}
	}
}
