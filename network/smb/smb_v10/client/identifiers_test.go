package client

import (
	"os"
	"testing"
)

// TestProcessIdentifierNeverNamesLockOwnerZero checks the conversion from an
// operating-system process identifier. Only the low 16 bits reach a lock range
// ([MS-CIFS] 2.2.4.32.1), so an identifier that is an exact multiple of 65536
// must not be carried through as-is.
func TestProcessIdentifierNeverNamesLockOwnerZero(t *testing.T) {
	tests := []struct {
		pid  int
		want uint32
	}{
		{1, 1},
		{4127, 4127},
		{0x1_0000, 0x1_0001},
		{0x2_0000, 0x2_0001},
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

// TestNextMIDIncrementsAndSkipsZero is the regression guard for requests that
// were all stamped with the same MID: consecutive requests on a connection must
// receive distinct identifiers, and zero is never handed out.
func TestNextMIDIncrementsAndSkipsZero(t *testing.T) {
	c := &Connection{}

	seen := make(map[uint16]bool)
	for i := 0; i < 4; i++ {
		mid := uint16(c.nextMID())
		if mid == 0 {
			t.Fatalf("request %d was given MID 0", i)
		}
		if seen[mid] {
			t.Fatalf("request %d reused MID %d", i, mid)
		}
		seen[mid] = true
	}
	if uint16(c.nextMID()) != 5 {
		t.Error("MIDs are not handed out in sequence")
	}

	// At the point the 16-bit field wraps, zero is stepped over rather than
	// emitted.
	c.messageIDCounter.Store(0xFFFF)
	if mid := uint16(c.nextMID()); mid != 1 {
		t.Errorf("MID after wrap = %d, want 1", mid)
	}
}
