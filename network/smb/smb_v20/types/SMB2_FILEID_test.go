package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

func TestSMB2_FILEID_MarshalUnmarshalRoundTrip(t *testing.T) {
	original := types.SMB2_FILEID{
		Persistent: 0x1122334455667788,
		Volatile:   0x99AABBCCDDEEFF00,
	}

	marshalled, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal() returned error: %v", err)
	}
	if len(marshalled) != types.SMB2_FILEID_SIZE {
		t.Fatalf("Marshal() produced %d bytes, want %d", len(marshalled), types.SMB2_FILEID_SIZE)
	}

	// Persistent and Volatile are each little-endian uint64.
	wantBytes := []byte{
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11,
		0x00, 0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99,
	}
	if !bytes.Equal(marshalled, wantBytes) {
		t.Errorf("Marshal() = % x, want % x", marshalled, wantBytes)
	}

	var decoded types.SMB2_FILEID
	n, err := decoded.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal() returned error: %v", err)
	}
	if n != types.SMB2_FILEID_SIZE {
		t.Errorf("Unmarshal() read %d bytes, want %d", n, types.SMB2_FILEID_SIZE)
	}
	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestSMB2_FILEID_UnmarshalShort(t *testing.T) {
	var f types.SMB2_FILEID
	if _, err := f.Unmarshal([]byte{0x00, 0x01, 0x02}); err == nil {
		t.Errorf("expected error unmarshalling a short buffer, got nil")
	}
}
