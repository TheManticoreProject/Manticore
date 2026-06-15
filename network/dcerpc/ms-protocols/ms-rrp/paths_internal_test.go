package ms_rrp

import (
	"errors"
	"fmt"
	"testing"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
)

func TestSplitRegistryPath(t *testing.T) {
	cases := []struct {
		in, root, sub string
	}{
		{`HKLM\Software\Microsoft`, "HKLM", `Software\Microsoft`},
		{`\HKEY_LOCAL_MACHINE\SYSTEM\`, "HKEY_LOCAL_MACHINE", "SYSTEM"},
		{`HKCU`, "HKCU", ""},
		{`  HKU\.DEFAULT  `, "HKU", ".DEFAULT"},
	}
	for _, c := range cases {
		root, sub := splitRegistryPath(c.in)
		if root != c.root || sub != c.sub {
			t.Errorf("splitRegistryPath(%q) = (%q, %q), want (%q, %q)", c.in, root, sub, c.root, c.sub)
		}
	}
}

func TestRegNameTerminatorCounted(t *testing.T) {
	// "Foo" is 3 wchars; with the appended NUL the counted byte Length must be 8.
	u := regName("Foo")
	if u.Length != 8 {
		t.Errorf("regName(\"Foo\").Length = %d, want 8 (3 chars + NUL, in bytes)", u.Length)
	}
}

func TestIsStatus(t *testing.T) {
	err := fmt.Errorf("BaseRegEnumKey failed: %s", winreg.StatusString(winreg.ErrorNoMoreItems))
	if !isStatus(err, winreg.ErrorNoMoreItems) {
		t.Error("isStatus did not match ERROR_NO_MORE_ITEMS")
	}
	if isStatus(err, winreg.ErrorAccessDenied) {
		t.Error("isStatus matched the wrong code")
	}
	if isStatus(nil, winreg.ErrorNoMoreItems) {
		t.Error("isStatus(nil) should be false")
	}
	if isStatus(errors.New("plain"), winreg.ErrorMoreData) {
		t.Error("isStatus matched an unrelated error")
	}
}

func TestUTF16RoundTrip(t *testing.T) {
	for _, s := range []string{"", "A", "Hello, World", "Ünîcödé", `C:\Windows`} {
		if got := decodeUTF16(encodeUTF16(s)); got != s {
			t.Errorf("utf16 round-trip %q = %q", s, got)
		}
	}
	// encodeUTF16 must append exactly one NUL terminator (2 octets).
	if n := len(encodeUTF16("AB")); n != 6 {
		t.Errorf("encodeUTF16(\"AB\") len = %d, want 6 (2 chars + NUL)", n)
	}
}
