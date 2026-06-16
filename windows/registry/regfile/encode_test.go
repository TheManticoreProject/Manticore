package regfile

import (
	"reflect"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/registry"
)

func TestFormatKeyHeader(t *testing.T) {
	const path = `HKEY_LOCAL_MACHINE\Software\Foo`
	if got, want := FormatKeyHeader(path, false), "["+path+"]"; got != want {
		t.Errorf("FormatKeyHeader = %q, want %q", got, want)
	}
	if got, want := FormatKeyHeader(path, true), "[-"+path+"]"; got != want {
		t.Errorf("FormatKeyHeader(delete) = %q, want %q", got, want)
	}
}

func TestFormatValueLine(t *testing.T) {
	cases := []struct {
		name  string
		value registry.Value
		want  string
	}{
		{"", registry.StringValue("default"), `@="default"`},
		{"Str", registry.StringValue(`C:\dir "q"`), `"Str"="C:\\dir \"q\""`},
		{"D", registry.DwordValue(0x1f), `"D"=dword:0000001f`},
		{"Q", registry.QwordValue(0x1f), `"Q"=hex(b):1f,00,00,00,00,00,00,00`},
		{"B", registry.BinaryValue([]byte{0xde, 0xad}), `"B"=hex:de,ad`},
		{"N", registry.NoneValue(nil), `"N"=hex(0):`},
		{"X", registry.Value{Type: 0x100, Data: []byte{0x01}}, `"X"=hex(100):01`},
	}
	for _, c := range cases {
		if got := FormatValueLine(c.name, c.value); got != c.want {
			t.Errorf("FormatValueLine(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestFormatDeleteValueLine(t *testing.T) {
	if got, want := FormatDeleteValueLine("Gone"), `"Gone"=-`; got != want {
		t.Errorf("FormatDeleteValueLine = %q, want %q", got, want)
	}
	if got, want := FormatDeleteValueLine(""), `@=-`; got != want {
		t.Errorf("FormatDeleteValueLine(default) = %q, want %q", got, want)
	}
}

func TestMarshalHeaderAndBOM(t *testing.T) {
	out := Marshal([]KeyBlock{{Path: `HKEY_CURRENT_USER\X`, Values: []ValueLine{{Name: "v", Value: registry.DwordValue(1)}}}})
	if len(out) < 2 || out[0] != 0xFF || out[1] != 0xFE {
		t.Fatalf("Marshal output missing UTF-16LE BOM: % x", out[:min(2, len(out))])
	}
	text := decodeToUTF8(out)
	if !strings.HasPrefix(text, Header+"\r\n\r\n") {
		t.Errorf("Marshal output does not start with header + blank line; got %q", text[:min(60, len(text))])
	}
}

func TestHexLineWrapping(t *testing.T) {
	data := make([]byte, 64) // long enough to force a continuation
	for i := range data {
		data[i] = byte(i)
	}
	line := FormatValueLine("Long", registry.Value{Type: registry.RegMultiSz, Data: data})
	if !strings.Contains(line, "\\\r\n") {
		t.Errorf("expected backslash continuation in wrapped hex line, got:\n%s", line)
	}
	// It must still round-trip through the folding decoder.
	v, _, err := parseValueData(strings.TrimPrefix(foldOne(line), `"Long"=`))
	if err != nil {
		t.Fatalf("parseValueData after fold: %v", err)
	}
	if !reflect.DeepEqual(v.Data, data) {
		t.Errorf("wrapped hex did not round-trip: got % x", v.Data)
	}
}

// foldOne runs a single multi-physical-line string through the continuation
// folder and returns the single logical line.
func foldOne(s string) string {
	folded := foldContinuations(splitLines(s))
	return strings.Join(folded, "")
}

// --- round-trip: Parse(Marshal(x)) == x for every value type ---

func TestRoundTrip(t *testing.T) {
	in := []KeyBlock{
		{
			Path: `HKEY_LOCAL_MACHINE\SOFTWARE\Manticore\RegFile`,
			Values: []ValueLine{
				{Name: "", Value: registry.StringValue("the default value")},
				{Name: "Str", Value: registry.StringValue(`path C:\windows\system32 and a "quote"`)},
				{Name: "Empty", Value: registry.StringValue("")},
				{Name: "Expand", Value: registry.ExpandStringValue("%SystemRoot%\\foo")},
				{Name: "Dword", Value: registry.DwordValue(0xdeadbeef)},
				{Name: "Qword", Value: registry.QwordValue(0x1122334455667788)},
				{Name: "Bin", Value: registry.BinaryValue([]byte{0x00, 0x01, 0x7f, 0x80, 0xff})},
				{Name: "Multi", Value: registry.MultiStringValue([]string{"alpha", "beta", "gamma gamma gamma"})},
				{Name: "None", Value: registry.NoneValue([]byte{0x01, 0x02})},
				{Name: "Custom", Value: registry.Value{Type: 0x77, Data: []byte{0xaa, 0xbb}}},
				{Name: "Doomed", Delete: true},
			},
		},
		{Path: `HKEY_CURRENT_USER\Software\Obsolete`, Delete: true},
	}

	out, err := Parse(Marshal(in))
	if err != nil {
		t.Fatalf("Parse(Marshal(in)) error: %v", err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Errorf("round-trip mismatch:\n got = %#v\nwant = %#v", out, in)
	}
}
