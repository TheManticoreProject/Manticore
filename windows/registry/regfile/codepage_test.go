package regfile

import (
	"os"
	"reflect"
	"testing"

	"golang.org/x/text/encoding/charmap"

	"github.com/TheManticoreProject/Manticore/windows/registry"
)

// The two files in testdata are regedit exports of the same key, made on a
// Windows Server 2012 R2 host (ANSI code page 1252) with "regedit /a" (REGEDIT4,
// ANSI) and "regedit /e" (version 5.00, UTF-16LE). The key holds a REG_SZ named
// "Café" with data "naïve € ü", a REG_EXPAND_SZ "%SystemRoot%\té" and a
// REG_BINARY 01 02 fe. In the REGEDIT4 file the names, the quoted string and
// the hex(2) data are all Windows-1252 bytes.

// TestParseRealREGEDIT4MatchesUnicodeExport checks the ANSI REGEDIT4 export
// decodes to exactly the same key blocks as the Unicode export of the same key.
func TestParseRealREGEDIT4MatchesUnicodeExport(t *testing.T) {
	ansiRaw, err := os.ReadFile("testdata/regedit4_cp1252.reg")
	if err != nil {
		t.Fatal(err)
	}
	unicodeRaw, err := os.ReadFile("testdata/regedit5_unicode.reg")
	if err != nil {
		t.Fatal(err)
	}
	ansi, err := Parse(ansiRaw)
	if err != nil {
		t.Fatalf("Parse(REGEDIT4): %v", err)
	}
	unicode, err := Parse(unicodeRaw)
	if err != nil {
		t.Fatalf("Parse(version 5): %v", err)
	}
	if !reflect.DeepEqual(ansi, unicode) {
		t.Fatalf("REGEDIT4 and Unicode exports differ:\n ansi    %#v\n unicode %#v", ansi, unicode)
	}

	values := map[string]registry.Value{}
	for _, v := range ansi[0].Values {
		values[v.Name] = v.Value
	}
	if s := values["Café"].String(); s != "naïve € ü" {
		t.Errorf(`"Café" = %q, want "naïve € ü"`, s)
	}
	exp := values["Exp"]
	if exp.Type != registry.RegExpandSz || exp.String() != `%SystemRoot%\té` {
		t.Errorf(`"Exp" = type %d %q, want REG_EXPAND_SZ "%%SystemRoot%%\té"`, exp.Type, exp.String())
	}
	if b := values["Bin"]; b.Type != registry.RegBinary || !reflect.DeepEqual(b.Data, []byte{1, 2, 0xfe}) {
		t.Errorf(`"Bin" = %#v`, b)
	}
}

// TestParseWithCodePage checks a REGEDIT4 file written in another code page
// (850, where 0x82 is "é" and 0x81 "ü") decodes through ParseWithCodePage.
func TestParseWithCodePage(t *testing.T) {
	src := []byte("REGEDIT4\r\n\r\n[HKEY_CURRENT_USER\\X]\r\n\"Caf\x82\"=\"\x81\"\r\n\"S\"=hex(1):82,00\r\n")
	blocks, err := ParseWithCodePage(src, charmap.CodePage850)
	if err != nil {
		t.Fatal(err)
	}
	if v := blocks[0].Values[0]; v.Name != "Café" || v.Value.String() != "ü" {
		t.Errorf("value = %q=%q, want Café=ü", v.Name, v.Value.String())
	}
	if s := blocks[0].Values[1].Value.String(); s != "é" {
		t.Errorf("hex(1) = %q, want é", s)
	}
}

// TestParseHexTypePrefixCaseInsensitive checks hand-edited prefixes such as
// "HEX(2):", "Hex:" and "DWORD:" are accepted like regedit's lower-case ones.
func TestParseHexTypePrefixCaseInsensitive(t *testing.T) {
	src := []byte("Windows Registry Editor Version 5.00\r\n\r\n[HKEY_CURRENT_USER\\X]\r\n" +
		"\"a\"=HEX(2):41,00,00,00\r\n\"b\"=Hex:01,02\r\n\"c\"=DWORD:0000000a\r\n\"d\"=hex(B):01,00,00,00,00,00,00,00\r\n")
	blocks, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	got := blocks[0].Values
	if len(got) != 4 || got[0].Value.Type != registry.RegExpandSz || got[1].Value.Type != registry.RegBinary ||
		got[3].Value.Type != registry.RegQword {
		t.Fatalf("values = %#v", got)
	}
	if n, ok := got[2].Value.Uint32(); !ok || n != 10 {
		t.Errorf("DWORD = %d (%v), want 10", n, ok)
	}
}

// TestParseVersion5AfterREGEDIT4 checks the REGEDIT4 code-page handling of hex
// data ends at a following version 5 header, whose hex(2) data is UTF-16LE.
func TestParseVersion5AfterREGEDIT4(t *testing.T) {
	src := []byte("REGEDIT4\r\n[HKEY_CURRENT_USER\\A]\r\n\"s\"=hex(2):e9,00\r\n" +
		"Windows Registry Editor Version 5.00\r\n[HKEY_CURRENT_USER\\B]\r\n\"s\"=hex(2):e9,00,00,00\r\n")
	blocks, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		if s := b.Values[0].Value.String(); s != "é" {
			t.Errorf("%s: hex(2) = %q, want é", b.Path, s)
		}
	}
}
