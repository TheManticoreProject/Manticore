package regfile

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/registry"
)

func TestParseUTF16LEWithBOM(t *testing.T) {
	src := Header + "\r\n\r\n" +
		"[HKEY_LOCAL_MACHINE\\Software\\Foo]\r\n" +
		"\"Answer\"=dword:0000002a\r\n"
	raw := encodeUTF16LEWithBOM(src)

	blocks, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Path != `HKEY_LOCAL_MACHINE\Software\Foo` {
		t.Fatalf("unexpected blocks: %#v", blocks)
	}
	v := blocks[0].Values[0]
	if n, ok := v.Value.Uint32(); v.Name != "Answer" || !ok || n != 42 {
		t.Errorf("value = %#v (n=%d ok=%v), want Answer=42", v, n, ok)
	}
}

func TestParseREGEDIT4UTF8(t *testing.T) {
	src := []byte("REGEDIT4\r\n\r\n" +
		"[HKEY_CURRENT_USER\\Console]\r\n" +
		"\"FaceName\"=\"Consolas\"\r\n")

	blocks, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if got := blocks[0].Values[0].Value.String(); got != "Consolas" {
		t.Errorf("FaceName = %q, want %q", got, "Consolas")
	}
}

func TestParseUTF8BOM(t *testing.T) {
	src := append([]byte{0xEF, 0xBB, 0xBF}, []byte(Header+"\n[HKEY_USERS\\X]\n\"v\"=\"y\"\n")...)
	blocks, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if blocks[0].Path != `HKEY_USERS\X` || blocks[0].Values[0].Value.String() != "y" {
		t.Errorf("unexpected blocks: %#v", blocks)
	}
}

func TestParseIgnoresCommentsAndBlanks(t *testing.T) {
	src := []byte(Header + "\n\n" +
		"; this is a comment\n" +
		"[HKEY_LOCAL_MACHINE\\K]\n" +
		"\n" +
		"; another comment\n" +
		"\"v\"=dword:00000001\n")
	blocks, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(blocks) != 1 || len(blocks[0].Values) != 1 {
		t.Fatalf("comments/blanks not ignored: %#v", blocks)
	}
}

func TestParseHexContinuationMultiSz(t *testing.T) {
	// REG_MULTI_SZ for ["ab","c"], wrapped across two lines with a backslash.
	src := []byte(Header + "\n" +
		"[HKEY_LOCAL_MACHINE\\K]\n" +
		"\"Multi\"=hex(7):61,00,62,00,00,00,\\\n" +
		"  63,00,00,00,00,00\n")
	blocks, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	v := blocks[0].Values[0].Value
	if v.Type != registry.RegMultiSz {
		t.Fatalf("type = %d, want REG_MULTI_SZ", v.Type)
	}
	if got := v.MultiString(); !reflect.DeepEqual(got, []string{"ab", "c"}) {
		t.Errorf("MultiString = %v, want [ab c]", got)
	}
}

func TestParseDeleteDirectives(t *testing.T) {
	src := []byte(Header + "\n" +
		"[-HKEY_LOCAL_MACHINE\\Gone]\n" +
		"[HKEY_LOCAL_MACHINE\\K]\n" +
		"\"Dead\"=-\n")
	blocks, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !blocks[0].Delete || blocks[0].Path != `HKEY_LOCAL_MACHINE\Gone` {
		t.Errorf("key delete not parsed: %#v", blocks[0])
	}
	if vl := blocks[1].Values[0]; !vl.Delete || vl.Name != "Dead" {
		t.Errorf("value delete not parsed: %#v", vl)
	}
}

func TestParseCommentWithTrailingBackslashKeepsNextLine(t *testing.T) {
	// A comment ending in a backslash must not swallow the following key line.
	src := []byte(Header + "\n" +
		"; see also C:\\Windows\\\n" +
		"[HKEY_LOCAL_MACHINE\\K]\n" +
		"\"v\"=dword:00000001\n")
	blocks, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Path != `HKEY_LOCAL_MACHINE\K` {
		t.Fatalf("comment continuation swallowed the key line: %#v", blocks)
	}
	if n, ok := blocks[0].Values[0].Value.Uint32(); !ok || n != 1 {
		t.Errorf("value lost: %#v", blocks[0].Values)
	}
}

func TestParseDefaultValueToken(t *testing.T) {
	blocks, err := Parse([]byte(Header + "\n[HKEY_X\\K]\n@=\"def\"\n"))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if vl := blocks[0].Values[0]; vl.Name != "" || vl.Value.String() != "def" {
		t.Errorf("default value = %#v, want Name=\"\" Value=def", vl)
	}
}

// --- helper unit tests ---

func TestScanQuoted(t *testing.T) {
	cases := []struct {
		in    string
		inner string
		rest  string
		ok    bool
	}{
		{`"abc"=x`, "abc", "=x", true},
		{`"a\\b\"c"`, `a\b"c`, "", true},
		{`"unterminated`, "", `"unterminated`, false},
		{`noquote`, "", "noquote", false},
	}
	for _, c := range cases {
		inner, rest, ok := scanQuoted(c.in)
		if inner != c.inner || rest != c.rest || ok != c.ok {
			t.Errorf("scanQuoted(%q) = (%q, %q, %v), want (%q, %q, %v)", c.in, inner, rest, ok, c.inner, c.rest, c.ok)
		}
	}
}

func TestParseHexBytes(t *testing.T) {
	got, err := parseHexBytes("de, ad,be,ef,")
	if err != nil {
		t.Fatalf("parseHexBytes error: %v", err)
	}
	if !reflect.DeepEqual(got, []byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Errorf("parseHexBytes = % x, want de ad be ef", got)
	}
	if got, _ := parseHexBytes("   "); got != nil {
		t.Errorf("parseHexBytes(empty) = % x, want nil", got)
	}
	if _, err := parseHexBytes("zz"); err == nil {
		t.Error("parseHexBytes(invalid) error = nil, want non-nil")
	}
}

func TestParseValueBeforeKeyReportsError(t *testing.T) {
	blocks, err := Parse([]byte(Header + "\n\"v\"=dword:00000001\n"))
	if err == nil {
		t.Error("expected error for value line before any key")
	}
	if len(blocks) != 0 {
		t.Errorf("expected no blocks, got %#v", blocks)
	}
}
