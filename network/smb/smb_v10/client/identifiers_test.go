package client

import (
	"os"
	"testing"
)

// TestProcessIdentifierFitsTheSMB1FieldAndNeverNamesLockOwnerZero checks that
// the host PID is reduced to the 16-bit value Windows carries, while reserving
// zero because that low half also names byte-range-lock ownership.
func TestProcessIdentifierFitsTheSMB1FieldAndNeverNamesLockOwnerZero(t *testing.T) {
	tests := []struct {
		pid  int
		want uint32
	}{
		{1, 1},
		{4127, 4127},
		{0x1_0000, 1},
		{0x2_0000, 1},
		{0x12_3456, 0x3456},
		{0, 1},
		{0xFFFF, 0xFFFF},
	}
	for _, tt := range tests {
		if got := uint32(processIdentifier(tt.pid)); got != tt.want {
			t.Errorf("processIdentifier(%d) = %#x, want %#x", tt.pid, got, tt.want)
		}
		if uint16(processIdentifier(tt.pid)) == 0 {
			t.Errorf("processIdentifier(%d) names lock owner 0", tt.pid)
		}
	}
}

// TestClientProcessIDIsThisProcess checks that the PID put on the wire is the
// identifier of the process sending the requests, not a placeholder.
func TestClientProcessIDIsThisProcess(t *testing.T) {
	if clientProcessID == 0 {
		t.Fatal("clientProcessID is 0")
	}
	if want := processIdentifier(os.Getpid()); clientProcessID != want {
		t.Errorf("clientProcessID = %#x, want %#x", uint32(clientProcessID), uint32(want))
	}
}

// TestNextMIDStartsAtZeroAndIncrements is the regression guard for requests
// that were all stamped with the same MID. Zero is a valid MID and is what a
// Windows client uses for its first request.
func TestNextMIDStartsAtZeroAndIncrements(t *testing.T) {
	c := &Connection{}

	for i := 0; i < 4; i++ {
		mid := uint16(c.nextMID())
		if mid != uint16(i) {
			t.Fatalf("request %d was given MID %d", i, mid)
		}
	}
	if uint16(c.nextMID()) != 4 {
		t.Error("MIDs are not handed out in sequence")
	}

	// The entire 16-bit field is valid: 0xFFFF is followed by zero.
	c.messageIDCounter.Store(0xFFFF)
	if mid := uint16(c.nextMID()); mid != 0xFFFF {
		t.Errorf("MID before wrap = %d, want 65535", mid)
	}
	if mid := uint16(c.nextMID()); mid != 0 {
		t.Errorf("MID after wrap = %d, want 0", mid)
	}
}
