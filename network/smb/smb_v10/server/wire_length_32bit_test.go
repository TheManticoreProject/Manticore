package server

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// The tests in this file send 32-bit wire values that are negative, or overflow
// a sum, as a 32-bit int. They are only meaningful under GOARCH=386, where the
// unguarded conversions panicked the request (dropping the connection); on a
// 64-bit build they still pin the status.

// TestSearchResumePositionPastInt32Ends asserts a resume position from the top
// half of the 32-bit range ends the search like any other position past the
// listing, rather than indexing before it.
func TestSearchResumePositionPastInt32Ends(t *testing.T) {
	_, client := searchServer(t, "ONE.TXT")

	resumeAt := func(position uint32) uint32 {
		key := types.SMB_RESUME_KEY{}
		binary.LittleEndian.PutUint32(key.ServerState[0:4], position)
		copy(key.ServerState[4:8], preNTSearchMagic[:])
		status, _ := sendSearch(t, client, `\*`, 10, &key)
		return status
	}

	want := resumeAt(100)
	for _, position := range []uint32{0x7FFFFFFF, 0x80000000, 0xFFFFFFFF} {
		if got := resumeAt(position); got != want {
			t.Errorf("resume position 0x%08X reported 0x%08X, want 0x%08X as for any position past the listing",
				position, got, want)
		}
	}
}

// TestNtTransactCreateNameLengthOverflowRefused asserts a NameLength that
// overflows the parameter block is refused as invalid rather than sliced.
func TestNtTransactCreateNameLengthOverflowRefused(t *testing.T) {
	for _, nameLength := range []uint32{0x7FFFFFF0, 0x80000000, 0xFFFFFFFF} {
		params := make([]byte, 64)
		binary.LittleEndian.PutUint32(params[40:44], nameLength)
		reassembly := &transactionReassembly{parameters: params}

		var status nt_status.NT_STATUS
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NameLength 0x%08X: handleNtTransactCreate panicked: %v", nameLength, r)
				}
			}()
			_, _, status = handleNtTransactCreate(&Connection{}, newRequest(0xA0), reassembly)
		}()
		if status != nt_status.NT_STATUS_INVALID_PARAMETER {
			t.Errorf("NameLength 0x%08X reported 0x%08X, want STATUS_INVALID_PARAMETER", nameLength, uint32(status))
		}
	}
}
