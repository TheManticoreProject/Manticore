package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestTreeConnectAndxResponseBaseWordCount verifies the base CIFS response still marshals
// with WordCount 0x03 and is not decoded as extended.
func TestTreeConnectAndxResponseBaseWordCount(t *testing.T) {
	resp := commands.NewTreeConnectAndxResponse()
	resp.OptionalSupport = types.USHORT(0x0001)

	marshalled, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if marshalled[0] != 0x03 {
		t.Errorf("base WordCount: got 0x%02x, want 0x03", marshalled[0])
	}

	var out commands.TreeConnectAndxResponse
	if _, err := out.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if out.Extended {
		t.Errorf("base response decoded as Extended")
	}
	if out.OptionalSupport != 0x0001 {
		t.Errorf("OptionalSupport round trip: got 0x%04x, want 0x0001", out.OptionalSupport)
	}
}

// TestTreeConnectAndxResponseExtended verifies the MS-SMB extended response form: the two
// ACCESS_MASK fields are appended after OptionalSupport, WordCount becomes 0x07, and the
// message round-trips.
func TestTreeConnectAndxResponseExtended(t *testing.T) {
	resp := commands.NewTreeConnectAndxResponse()
	resp.OptionalSupport = types.USHORT(0x0001)
	resp.Extended = true
	resp.MaximalShareAccessRights = types.ULONG(0x001F01FF)
	resp.GuestMaximalShareAccessRights = types.ULONG(0x00120089)

	marshalled, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if marshalled[0] != 0x07 {
		t.Errorf("extended WordCount: got 0x%02x, want 0x07", marshalled[0])
	}

	// Words block = 14 octets (AndX 4 + OptionalSupport 2 + two ACCESS_MASK 8); the 8
	// extended octets are the tail of the Words block (message octets [7:15]).
	extended := marshalled[1+6 : 1+14]
	want := make([]byte, 8)
	binary.LittleEndian.PutUint32(want[0:4], uint32(resp.MaximalShareAccessRights))
	binary.LittleEndian.PutUint32(want[4:8], uint32(resp.GuestMaximalShareAccessRights))
	if !bytes.Equal(extended, want) {
		t.Errorf("extended fields:\n got % x\nwant % x", extended, want)
	}

	var out commands.TreeConnectAndxResponse
	if _, err := out.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if !out.Extended {
		t.Fatalf("extended response not decoded as Extended")
	}
	if out.MaximalShareAccessRights != resp.MaximalShareAccessRights {
		t.Errorf("MaximalShareAccessRights round trip: got 0x%08x, want 0x%08x", out.MaximalShareAccessRights, resp.MaximalShareAccessRights)
	}
	if out.GuestMaximalShareAccessRights != resp.GuestMaximalShareAccessRights {
		t.Errorf("GuestMaximalShareAccessRights round trip: got 0x%08x, want 0x%08x", out.GuestMaximalShareAccessRights, resp.GuestMaximalShareAccessRights)
	}
}
