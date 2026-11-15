package server

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// TestResponseHeaderEchoesReserved checks that a sync response echoes the
// request's Reserved field (the legacy ProcessId; Windows clients send
// 0x0000FEFF). Zeroing it while the request carried 0xFEFF fingerprints the
// server, since Windows echoes the value.
func TestResponseHeaderEchoesReserved(t *testing.T) {
	const clientProcessID = 0x0000FEFF

	c := &Connection{credits: newCreditWindow()}

	req := &request{header: header.NewHeader()}
	req.header.MessageId = 1
	req.header.Credit = 1
	req.header.Reserved = clientProcessID

	resp := &response{}
	h := c.responseHeader(req, resp, nt_status.NT_STATUS_SUCCESS)

	if h.Reserved != clientProcessID {
		t.Fatalf("response Reserved = %#x, want %#x", h.Reserved, clientProcessID)
	}

	// Confirm it lands on the wire: the sync header carries Reserved at bytes
	// 32..35 (little-endian).
	wire, err := h.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint32(wire[32:36]); got != clientProcessID {
		t.Fatalf("wire Reserved (bytes 32..35) = %#x, want %#x", got, clientProcessID)
	}
}
