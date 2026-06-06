package commands

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestReadAndxResponseDataRoundTrip verifies that ReadAndxResponse can both
// serialize the bytes read from a file and recover them on Unmarshal, with
// DataLength/DataOffset set consistently.
func TestReadAndxResponseDataRoundTrip(t *testing.T) {
	payload := []byte("the quick brown fox\x00\x01\x02jumps over")

	resp := NewReadAndxResponse()
	resp.Available = types.USHORT(0)
	resp.Data = []types.UCHAR(payload)

	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if int(resp.DataLength) != len(payload) {
		t.Errorf("expected DataLength %d after Marshal, got %d", len(payload), resp.DataLength)
	}
	if resp.DataOffset != readAndxResponseDataOffset {
		t.Errorf("expected DataOffset %d after Marshal, got %d", readAndxResponseDataOffset, resp.DataOffset)
	}

	decoded := NewReadAndxResponse()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if int(decoded.DataLength) != len(payload) {
		t.Errorf("expected decoded DataLength %d, got %d", len(payload), decoded.DataLength)
	}
	if !bytes.Equal([]byte(decoded.Data), payload) {
		t.Errorf("decoded data mismatch:\n want %v\n got  %v", payload, []byte(decoded.Data))
	}
}

// TestReadAndxResponseEmptyData verifies that an end-of-file read (DataLength 0)
// round-trips to an empty Data slice without error.
func TestReadAndxResponseEmptyData(t *testing.T) {
	resp := NewReadAndxResponse()

	raw, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	decoded := NewReadAndxResponse()
	if _, err := decoded.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(decoded.Data) != 0 {
		t.Errorf("expected empty Data, got %d bytes", len(decoded.Data))
	}
	if decoded.DataLength != 0 {
		t.Errorf("expected DataLength 0, got %d", decoded.DataLength)
	}
}
