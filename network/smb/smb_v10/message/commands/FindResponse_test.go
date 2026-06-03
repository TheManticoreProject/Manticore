package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestFindResponseDataBlockFraming verifies that SMB_COM_FIND Response emits the
// SMB_Data block per [MS-CIFS] 2.2.4.59.2: a BufferFormat byte (0x05) followed by
// a 2-byte little-endian DataLength, then the DirectoryInformationData array.
// Previously the framing bytes were omitted and only the raw records were written.
func TestFindResponseDataBlockFraming(t *testing.T) {
	out := commands.NewFindResponse()
	out.Count = types.USHORT(0)

	marshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Layout: WordCount(1)=0x01, Count(2 LE), ByteCount(2 LE)=0x0003,
	// then the data block: BufferFormat(0x05) DataLength(00 00).
	// With no directory entries, the last three bytes must be the empty
	// variable-block framing 05 00 00.
	n := len(marshalled)
	if n < 3 || marshalled[n-3] != 0x05 || marshalled[n-2] != 0x00 || marshalled[n-1] != 0x00 {
		t.Fatalf("data block framing missing/incorrect; full = % x", marshalled)
	}
}
