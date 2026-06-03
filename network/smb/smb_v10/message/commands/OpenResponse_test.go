package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestOpenResponseWireFormat verifies that SMB_COM_OPEN Response marshals its
// parameter block per [MS-CIFS] 2.2.4.4.2: WordCount = 0x07 (14 bytes of Words),
// all integers little-endian, and LastModified as a 4-byte UTIME (not an 8-byte
// FILETIME).
func TestOpenResponseWireFormat(t *testing.T) {
	out := commands.NewOpenResponse()
	out.FID = types.USHORT(0x1234)
	out.LastModified = types.ULONG(0x11223344)
	out.FileSize = types.ULONG(0xAABBCCDD)
	out.AccessMode = types.USHORT(0x0102)

	marshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Expected wire bytes:
	//   WordCount    = 0x07
	//   FID          = 34 12          (0x1234, little-endian)
	//   FileAttrs    = 00 00
	//   LastModified = 44 33 22 11    (0x11223344, 4-byte UTIME, little-endian)
	//   FileSize     = dd cc bb aa    (0xAABBCCDD, little-endian)
	//   AccessMode   = 02 01          (0x0102, little-endian)
	//   ByteCount    = 00 00
	want := []byte{
		0x07,
		0x34, 0x12,
		0x00, 0x00,
		0x44, 0x33, 0x22, 0x11,
		0xdd, 0xcc, 0xbb, 0xaa,
		0x02, 0x01,
		0x00, 0x00,
	}
	if !bytes.Equal(marshalled, want) {
		t.Errorf("Marshal mismatch:\n got = % x\nwant = % x", marshalled, want)
	}
}
