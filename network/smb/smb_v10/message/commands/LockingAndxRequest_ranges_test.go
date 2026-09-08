package commands

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// lockingRoundTrip marshals a request and decodes it back, returning the decoded
// copy. The AndX block and framing come along for free, which is what a real
// request carries.
func lockingRoundTrip(t *testing.T, request *LockingAndxRequest) *LockingAndxRequest {
	t.Helper()

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	decoded := NewLockingAndxRequest()
	if _, err := decoded.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return decoded
}

// TestLockingAndxRangesRoundTripInBothFormats asserts a lock range survives a
// round trip in each of the two wire formats, and that each format occupies the
// width [MS-CIFS] 2.2.4.32.1 gives it.
//
// The width is the whole point: the formats are 10 and 20 bytes, so decoding an
// array in the wrong one does not fail cleanly — it assembles ranges out of
// adjacent entries and locks somewhere the client never asked for.
func TestLockingAndxRangesRoundTripInBothFormats(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		typeOfLock uint8
		entrySize  int
	}{
		{name: "32-bit ranges", typeOfLock: 0x00, entrySize: 10},
		{name: "64-bit ranges", typeOfLock: LockingAndxLargeFiles, entrySize: 20},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := NewLockingAndxRequest()
			request.FID = types.USHORT(0x0042)
			request.TypeOfLock = types.UCHAR(testCase.typeOfLock)
			request.Timeout = types.ULONG(0)

			locks := []types.LOCKING_ANDX_RANGE64{
				{PID: 0x1111, ByteOffsetLow: 0x00001000, LengthInBytesLow: 0x00000200},
				{PID: 0x2222, ByteOffsetLow: 0xFFFFFFF0, LengthInBytesLow: 0x00000010},
			}
			unlocks := []types.LOCKING_ANDX_RANGE64{
				{PID: 0x3333, ByteOffsetLow: 0x00000040, LengthInBytesLow: 0x00000004},
			}

			request.Locks = locks
			request.NumberOfRequestedLocks = types.USHORT(len(locks))
			request.Unlocks = unlocks
			request.NumberOfRequestedUnlocks = types.USHORT(len(unlocks))

			marshalled, err := request.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// WordCount(1) + 8 words + ByteCount(2), then the arrays.
			wantData := testCase.entrySize * (len(locks) + len(unlocks))
			gotData := len(marshalled) - (1 + 2*8 + 2)
			if gotData != wantData {
				t.Fatalf("the arrays occupy %d bytes, want %d (%d entries of %d)",
					gotData, wantData, len(locks)+len(unlocks), testCase.entrySize)
			}

			decoded := NewLockingAndxRequest()
			if _, err := decoded.Unmarshal(marshalled); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if got := len(decoded.Locks); got != len(locks) {
				t.Fatalf("decoded %d locks, want %d", got, len(locks))
			}
			if got := len(decoded.Unlocks); got != len(unlocks) {
				t.Fatalf("decoded %d unlocks, want %d", got, len(unlocks))
			}

			for index, want := range locks {
				got := decoded.Locks[index]
				if got.PID != want.PID || got.ByteOffsetLow != want.ByteOffsetLow ||
					got.LengthInBytesLow != want.LengthInBytesLow ||
					got.ByteOffsetHigh != want.ByteOffsetHigh ||
					got.LengthInBytesHigh != want.LengthInBytesHigh {
					t.Errorf("lock %d round-tripped to %+v, want %+v", index, got, want)
				}
			}
			for index, want := range unlocks {
				got := decoded.Unlocks[index]
				if got.PID != want.PID || got.ByteOffsetLow != want.ByteOffsetLow ||
					got.LengthInBytesLow != want.LengthInBytesLow {
					t.Errorf("unlock %d round-tripped to %+v, want %+v", index, got, want)
				}
			}
		})
	}
}

// TestLockingAndxLargeRangeNeedsTheLargeFilesBit asserts a range that does not fit
// in 32 bits is refused rather than truncated when the 32-bit format is declared.
//
// Truncating would name a different part of the file and report success, which is
// worse than refusing: the caller would believe it holds a lock it does not.
func TestLockingAndxLargeRangeNeedsTheLargeFilesBit(t *testing.T) {
	request := NewLockingAndxRequest()
	request.FID = types.USHORT(0x0042)
	request.TypeOfLock = types.UCHAR(0x00)
	request.Locks = []types.LOCKING_ANDX_RANGE64{
		{PID: 0x1111, ByteOffsetHigh: 0x00000001, ByteOffsetLow: 0x00000000, LengthInBytesLow: 0x10},
	}
	request.NumberOfRequestedLocks = types.USHORT(1)

	if _, err := request.Marshal(); err == nil {
		t.Fatal("a range past 4 GiB marshalled into the 32-bit format, want an error")
	}

	// With the bit set the same range is expressible. A fresh request is built
	// rather than the previous one reused: Marshal appends to a command's
	// parameter and data blocks, so marshalling one twice does not produce the
	// same bytes twice.
	wide := NewLockingAndxRequest()
	wide.FID = types.USHORT(0x0042)
	wide.TypeOfLock = types.UCHAR(LockingAndxLargeFiles)
	wide.Locks = []types.LOCKING_ANDX_RANGE64{
		{PID: 0x1111, ByteOffsetHigh: 0x00000001, ByteOffsetLow: 0x00000000, LengthInBytesLow: 0x10},
	}
	wide.NumberOfRequestedLocks = types.USHORT(1)

	decoded := lockingRoundTrip(t, wide)
	if got := len(decoded.Locks); got != 1 {
		t.Fatalf("decoded %d locks, want 1", got)
	}
	if got := decoded.Locks[0].ByteOffsetHigh; got != 0x00000001 {
		t.Errorf("the high offset word round-tripped to 0x%08X, want 0x00000001", uint32(got))
	}
}

// TestLockingAndxTruncatedRangeArrayIsRefused asserts an array shorter than the
// count claims is reported rather than read past.
func TestLockingAndxTruncatedRangeArrayIsRefused(t *testing.T) {
	request := NewLockingAndxRequest()
	request.FID = types.USHORT(0x0042)
	request.TypeOfLock = types.UCHAR(0x00)
	request.Locks = []types.LOCKING_ANDX_RANGE64{
		{PID: 0x1111, ByteOffsetLow: 0x10, LengthInBytesLow: 0x10},
	}
	// One entry is sent, two are claimed.
	request.NumberOfRequestedLocks = types.USHORT(2)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	decoded := NewLockingAndxRequest()
	if _, err := decoded.Unmarshal(marshalled); err == nil {
		t.Fatal("a request claiming more ranges than it carries decoded, want an error")
	}
}
