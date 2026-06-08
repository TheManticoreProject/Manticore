package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestNtCreateAndxResponseBaseWordCount verifies the base CIFS response still marshals
// with WordCount 0x22 (34 words) and that the extended fields are absent.
func TestNtCreateAndxResponseBaseWordCount(t *testing.T) {
	resp := commands.NewNtCreateAndxResponse()
	resp.FID = types.USHORT(0x4321)

	marshalled, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if marshalled[0] != 0x22 {
		t.Errorf("base WordCount: got 0x%02x, want 0x22", marshalled[0])
	}

	var out commands.NtCreateAndxResponse
	if _, err := out.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if out.Extended {
		t.Errorf("base response decoded as Extended")
	}
	if out.FID != 0x4321 {
		t.Errorf("FID round trip: got 0x%04x, want 0x4321", out.FID)
	}
}

// TestNtCreateAndxResponseExtended verifies the MS-SMB extended response form: the four
// extended fields are appended in order (VolumeGUID, FileId, MaximalAccessRights,
// GuestMaximalAccessRights), WordCount becomes 0x32 (50 words, self-consistent with the
// 100-octet Words block), and the message round-trips.
func TestNtCreateAndxResponseExtended(t *testing.T) {
	resp := commands.NewNtCreateAndxResponse()
	resp.FID = types.USHORT(0x4321)
	resp.Extended = true
	resp.VolumeGUID = [16]byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}
	resp.FileId = 0x1122334455667788
	resp.MaximalAccessRights = types.ULONG(0x001F01FF)
	resp.GuestMaximalAccessRights = types.ULONG(0x00120089)

	marshalled, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// WordCount = 50 words (0x32); Words block = 100 octets; ByteCount = 2 octets.
	if marshalled[0] != 0x32 {
		t.Errorf("extended WordCount: got 0x%02x, want 0x32", marshalled[0])
	}
	if len(marshalled) != 1+100+2 {
		t.Fatalf("extended length: got %d, want %d", len(marshalled), 1+100+2)
	}

	// The 32 extended octets are the tail of the Words block (octets [69:101] of the
	// message: 1 WordCount + 68 base + 32 extended).
	extended := marshalled[1+68 : 1+100]
	want := []byte{}
	want = append(want, resp.VolumeGUID[:]...)
	b8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b8, resp.FileId)
	want = append(want, b8...)
	b4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(b4, uint32(resp.MaximalAccessRights))
	want = append(want, b4...)
	b4 = make([]byte, 4)
	binary.LittleEndian.PutUint32(b4, uint32(resp.GuestMaximalAccessRights))
	want = append(want, b4...)
	if !bytes.Equal(extended, want) {
		t.Errorf("extended fields:\n got % x\nwant % x", extended, want)
	}

	var out commands.NtCreateAndxResponse
	if _, err := out.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !out.Extended {
		t.Fatalf("extended response not decoded as Extended")
	}
	if out.VolumeGUID != resp.VolumeGUID {
		t.Errorf("VolumeGUID round trip: got % x, want % x", out.VolumeGUID, resp.VolumeGUID)
	}
	if out.FileId != resp.FileId {
		t.Errorf("FileId round trip: got 0x%016x, want 0x%016x", out.FileId, resp.FileId)
	}
	if out.MaximalAccessRights != resp.MaximalAccessRights {
		t.Errorf("MaximalAccessRights round trip: got 0x%08x, want 0x%08x", out.MaximalAccessRights, resp.MaximalAccessRights)
	}
	if out.GuestMaximalAccessRights != resp.GuestMaximalAccessRights {
		t.Errorf("GuestMaximalAccessRights round trip: got 0x%08x, want 0x%08x", out.GuestMaximalAccessRights, resp.GuestMaximalAccessRights)
	}
}
