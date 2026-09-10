package client_test

import (
	"path/filepath"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/internal/wirediff"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
)

// TestParsesRecordedWindowsSMB1Responses runs every message a real Windows SMB
// 1.0 server sent through this client's message parser.
//
// The value is in where the bytes came from. Pairing this repository's client
// with its own server cannot detect a wire detail both halves get wrong, and
// that gap has produced real defects; these bytes were produced by Windows.
func TestParsesRecordedWindowsSMB1Responses(t *testing.T) {
	corpus, err := wirediff.Load(filepath.Join("..", "..", "testdata", "reference", "smb1-server2012r2.json"))
	if err != nil {
		t.Fatalf("loading corpus: %v", err)
	}

	responses, err := corpus.Messages(wirediff.FromServer)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}
	if len(responses) == 0 {
		t.Fatal("corpus records no server messages")
	}

	for i, raw := range responses {
		msg := message.NewMessage()
		if _, err := msg.Header.Unmarshal(raw); err != nil {
			t.Errorf("response %d (%d bytes): header did not parse: %v", i, len(raw), err)
			continue
		}
		if msg.Header.Protocol != [4]byte{0xFF, 'S', 'M', 'B'} {
			t.Errorf("response %d does not carry the SMB1 protocol id: % x", i, raw[:4])
		}
		if !msg.Header.IsResponse() {
			t.Errorf("response %d does not have SMB_FLAGS_REPLY set", i)
		}
		// Every response in this exchange was carried with Unicode strings, which
		// is what the client now negotiates.
		if !msg.Header.Flags2.IsUnicode() {
			t.Errorf("response %d (command %#02x) did not set SMB_FLAGS2_UNICODE", i, uint8(msg.Header.Command))
		}
	}
	t.Logf("parsed %d responses recorded from %s", len(responses), corpus.Peer)
}

// TestRecordedSMB1RequestsCarryUnicode checks the corpus against the defect it
// was recorded after: before the fix, every request after authentication
// cleared SMB_FLAGS2_UNICODE and carried paths as OEM, which silently corrupted
// any non-ASCII name on the server.
func TestRecordedSMB1RequestsCarryUnicode(t *testing.T) {
	corpus, err := wirediff.Load(filepath.Join("..", "..", "testdata", "reference", "smb1-server2012r2.json"))
	if err != nil {
		t.Fatalf("loading corpus: %v", err)
	}

	requests, err := corpus.Messages(wirediff.FromClient)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}

	for i, raw := range requests {
		msg := message.NewMessage()
		if _, err := msg.Header.Unmarshal(raw); err != nil {
			t.Fatalf("request %d: header did not parse: %v", i, err)
		}
		if !msg.Header.Flags2.IsUnicode() {
			t.Errorf("request %d (command %#02x) cleared SMB_FLAGS2_UNICODE; paths would go out as OEM",
				i, uint8(msg.Header.Command))
		}
	}
	t.Logf("all %d recorded requests carry SMB_FLAGS2_UNICODE", len(requests))
}
