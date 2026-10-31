package message_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// TestUnmarshalDoesNotAliasInputBuffer verifies that a decoded message does not
// keep a reference to the caller's buffer. The client and server read loops
// reuse a single receive buffer, so the typed RDATA accessors of a message that
// was already returned or dispatched must keep decoding the packet it came
// from, not whatever datagram overwrote the buffer afterwards.
func TestUnmarshalDoesNotAliasInputBuffer(t *testing.T) {
	// Response for "victim" PTR whose answer RDATA is a compression pointer
	// back to the question name at offset 12.
	buf := []byte{
		0x12, 0x34, 0x80, 0x00, // ID, flags (QR)
		0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // QD=1, AN=1
		6, 'v', 'i', 'c', 't', 'i', 'm', 0, // question name
		0x00, 0x0c, 0x00, 0x01, // PTR, IN
		0xC0, 12, // answer name -> question name
		0x00, 0x0c, 0x00, 0x01, 0x00, 0x00, 0x00, 0x1e, // PTR, IN, TTL 30
		0x00, 0x02, 0xC0, 12, // RDLENGTH 2, RDATA -> question name
	}

	var msg message.Message
	if _, err := msg.Unmarshal(buf); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	before, err := msg.Answers[0].AsPTR()
	if err != nil {
		t.Fatalf("AsPTR failed: %v", err)
	}
	if before != "victim" {
		t.Fatalf("AsPTR = %q, want %q", before, "victim")
	}

	// Simulate the read loop reusing the buffer for the next datagram.
	copy(buf[13:19], "ZZZZZZ")

	after, err := msg.Answers[0].AsPTR()
	if err != nil {
		t.Fatalf("AsPTR after buffer reuse failed: %v", err)
	}
	if after != before {
		t.Errorf("AsPTR changed from %q to %q after the input buffer was overwritten", before, after)
	}
}
