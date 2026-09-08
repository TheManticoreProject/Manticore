package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestNewSMB_RESUME_KEY asserts a fresh key is zeroed.
func TestNewSMB_RESUME_KEY(t *testing.T) {
	resumeKey := types.NewSMB_RESUME_KEY()
	if resumeKey == nil {
		t.Fatal("NewSMB_RESUME_KEY returned nil")
	}
	if resumeKey.Reserved != 0 {
		t.Errorf("Reserved is %d, want 0", resumeKey.Reserved)
	}
	if resumeKey.ServerState != [16]types.UCHAR{} {
		t.Errorf("ServerState is %v, want zeroes", resumeKey.ServerState)
	}
	if resumeKey.ClientState != [4]types.UCHAR{} {
		t.Errorf("ClientState is %v, want zeroes", resumeKey.ClientState)
	}
}

// TestSMB_RESUME_KEY_IsTwentyOneBareBytes asserts the key marshals its three
// fields and nothing else.
//
// It used to embed an SMB_STRING and emit a buffer-format byte and a length ahead
// of them, making it 24 bytes. Those two fields are real, but [MS-CIFS] section
// 2.2.4.58.1 places them in the SMB_COM_SEARCH *request's* data block: inside an
// SMB_DIRECTORY_INFORMATION entry the key is bare. One marshaller cannot produce
// both shapes, so the wrapper moved to the request and this stayed fixed-size.
func TestSMB_RESUME_KEY_IsTwentyOneBareBytes(t *testing.T) {
	resumeKey := &types.SMB_RESUME_KEY{
		Reserved:    0x11,
		ServerState: [16]types.UCHAR{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		ClientState: [4]types.UCHAR{0xAA, 0xBB, 0xCC, 0xDD},
	}

	marshalled, err := resumeKey.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if len(marshalled) != types.SMB_RESUME_KEY_SIZE {
		t.Fatalf("the key is %d bytes, want %d", len(marshalled), types.SMB_RESUME_KEY_SIZE)
	}

	want := append([]byte{0x11}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}...)
	want = append(want, 0xAA, 0xBB, 0xCC, 0xDD)
	if !bytes.Equal(marshalled, want) {
		t.Errorf("the key marshalled to % x, want % x", marshalled, want)
	}
}

// TestSMB_RESUME_KEY_RoundTrip asserts the key decodes what it encodes and
// consumes exactly its own length, which is what lets the field that follows it
// be found.
func TestSMB_RESUME_KEY_RoundTrip(t *testing.T) {
	resumeKey := &types.SMB_RESUME_KEY{
		Reserved:    0x7F,
		ServerState: [16]types.UCHAR{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		ClientState: [4]types.UCHAR{9, 8, 7, 6},
	}

	marshalled, err := resumeKey.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// A trailing byte stands in for the field that follows the key in a real
	// entry: consuming it would be the failure this guards against.
	withTrailer := append(append([]byte{}, marshalled...), 0xEE)

	decoded := types.NewSMB_RESUME_KEY()
	read, err := decoded.Unmarshal(withTrailer)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if read != types.SMB_RESUME_KEY_SIZE {
		t.Fatalf("Unmarshal consumed %d bytes, want %d", read, types.SMB_RESUME_KEY_SIZE)
	}
	if *decoded != *resumeKey {
		t.Errorf("the key round-tripped to %+v, want %+v", decoded, resumeKey)
	}
}

// TestSMB_RESUME_KEY_ShortDataIsRefused asserts a buffer that cannot hold the key
// is reported rather than read past.
func TestSMB_RESUME_KEY_ShortDataIsRefused(t *testing.T) {
	for _, length := range []int{0, 1, types.SMB_RESUME_KEY_SIZE - 1} {
		decoded := types.NewSMB_RESUME_KEY()
		if _, err := decoded.Unmarshal(make([]byte, length)); err == nil {
			t.Errorf("a %d-byte buffer decoded as a resume key, want an error", length)
		}
	}
}
