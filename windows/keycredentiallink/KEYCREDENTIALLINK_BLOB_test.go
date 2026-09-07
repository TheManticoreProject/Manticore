package keycredentiallink_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink"
)

// A blob shorter than the four-byte version field is malformed, and the value comes
// from a directory or a file, so it has to be reported as an error rather than
// slicing past the end of the buffer.
func TestKEYCREDENTIALLINK_BLOB_Unmarshal_TruncatedVersion(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "nil", data: nil},
		{name: "one byte", data: []byte{0x00}},
		{name: "three bytes", data: []byte{0x00, 0x02, 0x00}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blob := keycredentiallink.KEYCREDENTIALLINK_BLOB{}

			if _, err := blob.Unmarshal(test.data); err == nil {
				t.Errorf("Unmarshal(%d bytes) returned no error, want one", len(test.data))
			}
		})
	}
}

// A blob carrying only a version parses to zero entries: the version is complete, and
// the absence of the entries the specification marks mandatory is the caller's to
// handle, not a parse failure.
func TestKEYCREDENTIALLINK_BLOB_Unmarshal_VersionOnly(t *testing.T) {
	blob := keycredentiallink.KEYCREDENTIALLINK_BLOB{}

	bytesRead, err := blob.Unmarshal([]byte{0x00, 0x02, 0x00, 0x00})
	if err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}

	if bytesRead != 4 {
		t.Errorf("bytesRead = %d, want 4", bytesRead)
	}

	if len(blob.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(blob.Entries))
	}
}
