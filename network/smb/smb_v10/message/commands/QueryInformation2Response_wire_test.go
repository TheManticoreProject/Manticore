package commands

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestQueryInformation2ResponseParameterBlockIsElevenWords asserts the response
// occupies the parameter block [MS-CIFS] section 2.2.4.35.2 gives it.
//
// WordCount is 11: CreateDate(2) CreationTime(2) LastAccessDate(2)
// LastAccessTime(2) LastWriteDate(2) LastWriteTime(2) FileDataSize(4)
// FileAllocationSize(4) FileAttributes(2), which is 22 bytes. A time field of the
// wrong width does not fail loudly — it pushes every field behind it out of place,
// so the length is the assertion that catches it.
func TestQueryInformation2ResponseParameterBlockIsElevenWords(t *testing.T) {
	response := NewQueryInformation2Response()
	response.CreateDate = *types.NewSMB_DATEFromDate(2001, 2, 3)
	response.CreationTime = *types.NewSMB_TIME_DOSFromTime(4, 5, 6)
	response.LastAccessDate = *types.NewSMB_DATEFromDate(2002, 3, 4)
	response.LastAccessTime = *types.NewSMB_TIME_DOSFromTime(7, 8, 10)
	response.LastWriteDate = *types.NewSMB_DATEFromDate(2003, 4, 5)
	response.LastWriteTime = *types.NewSMB_TIME_DOSFromTime(11, 12, 14)
	response.FileDataSize = types.ULONG(0x11223344)
	response.FileAllocationSize = types.ULONG(0x55667788)

	marshalled, err := response.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// WordCount(1) + Words + ByteCount(2).
	const wantWords = 11
	if got := int(marshalled[0]); got != wantWords {
		t.Errorf("WordCount is %d, want %d", got, wantWords)
	}
	if want := 1 + 2*wantWords + 2; len(marshalled) != want {
		t.Fatalf("the command is %d bytes, want %d", len(marshalled), want)
	}
}

// TestQueryInformation2ResponseRoundTrips asserts the response decodes what it
// encodes.
//
// It could not before: each time field marshalled eight bytes and was handed a
// two-byte window on the way back, so Unmarshal failed on its own output.
func TestQueryInformation2ResponseRoundTrips(t *testing.T) {
	response := NewQueryInformation2Response()
	response.CreateDate = *types.NewSMB_DATEFromDate(2001, 2, 3)
	response.CreationTime = *types.NewSMB_TIME_DOSFromTime(4, 5, 6)
	response.LastAccessDate = *types.NewSMB_DATEFromDate(2002, 3, 4)
	response.LastAccessTime = *types.NewSMB_TIME_DOSFromTime(7, 8, 10)
	response.LastWriteDate = *types.NewSMB_DATEFromDate(2003, 4, 5)
	response.LastWriteTime = *types.NewSMB_TIME_DOSFromTime(11, 12, 14)
	response.FileDataSize = types.ULONG(0x11223344)
	response.FileAllocationSize = types.ULONG(0x55667788)

	marshalled, err := response.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	decoded := NewQueryInformation2Response()
	if _, err := decoded.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.FileDataSize != response.FileDataSize {
		t.Errorf("FileDataSize round-tripped to 0x%08X, want 0x%08X",
			uint32(decoded.FileDataSize), uint32(response.FileDataSize))
	}
	if decoded.FileAllocationSize != response.FileAllocationSize {
		t.Errorf("FileAllocationSize round-tripped to 0x%08X, want 0x%08X",
			uint32(decoded.FileAllocationSize), uint32(response.FileAllocationSize))
	}
	if decoded.LastWriteTime != response.LastWriteTime {
		t.Errorf("LastWriteTime round-tripped to %+v, want %+v",
			decoded.LastWriteTime, response.LastWriteTime)
	}
	if decoded.LastWriteDate != response.LastWriteDate {
		t.Errorf("LastWriteDate round-tripped to %+v, want %+v",
			decoded.LastWriteDate, response.LastWriteDate)
	}
	if decoded.CreationTime != response.CreationTime {
		t.Errorf("CreationTime round-tripped to %+v, want %+v",
			decoded.CreationTime, response.CreationTime)
	}
}
