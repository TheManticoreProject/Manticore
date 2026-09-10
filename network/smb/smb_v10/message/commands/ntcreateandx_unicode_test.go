package commands

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestNtCreateAndxUnicodeNameRoundTrips guards the alignment defect: a Unicode
// FileName is 16-bit aligned relative to the SMB header ([MS-CIFS] 2.2.4.64.1),
// and this command's data block starts at the odd offset 83, so the name needs a
// preceding pad byte and a two-byte terminator.
//
// Unmarshal already expected both. Marshal emitted neither, so the command could
// parse a Unicode name but not produce one, and a peer read every character one
// byte out of phase — "batched.txt" arrived as "戀愀琀挀栀攀搀⸀琀砀琀".
func TestNtCreateAndxUnicodeNameRoundTrips(t *testing.T) {
	for _, name := range []string{
		`\plain.txt`,
		`\unicode_éàü`,
		`\dir\nested_ß.txt`,
		`\`,
	} {
		t.Run(name, func(t *testing.T) {
			req := NewNtCreateAndxRequest()
			req.SetUnicode(true)
			if err := req.FileName.SetStringWithEncoding(name, true); err != nil {
				t.Fatalf("SetStringWithEncoding: %v", err)
			}
			req.NameLength = types.USHORT(len(req.FileName.Buffer))

			raw, err := req.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			parsed := NewNtCreateAndxRequest()
			parsed.SetUnicode(true)
			if _, err := parsed.Unmarshal(raw); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			if got := utf16.DecodeUTF16LE([]byte(parsed.FileName.Buffer)); got != name {
				t.Errorf("FileName round-tripped as %q, want %q", got, name)
			}
		})
	}
}

// TestNtCreateAndxOEMNameUnchanged pins the non-Unicode path, so the alignment
// work cannot leak a stray pad byte into an OEM message.
func TestNtCreateAndxOEMNameUnchanged(t *testing.T) {
	const name = `\plain.txt`

	req := NewNtCreateAndxRequest()
	req.SetUnicode(false)
	if err := req.FileName.SetStringWithEncoding(name, false); err != nil {
		t.Fatalf("SetStringWithEncoding: %v", err)
	}
	req.NameLength = types.USHORT(len(req.FileName.Buffer))

	raw, err := req.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	parsed := NewNtCreateAndxRequest()
	parsed.SetUnicode(false)
	if _, err := parsed.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := string(parsed.FileName.Buffer); got != name {
		t.Errorf("FileName round-tripped as %q, want %q", got, name)
	}
}
