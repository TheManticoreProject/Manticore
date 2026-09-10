package client

import (
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/internal/wirediff"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
)

// referenceCorpus loads a recorded exchange from the shared testdata directory.
func referenceCorpus(t *testing.T, name string) *wirediff.Corpus {
	t.Helper()

	corpus, err := wirediff.Load(filepath.Join("..", "..", "testdata", "reference", name))
	if err != nil {
		t.Fatalf("loading corpus: %v", err)
	}
	return corpus
}

// TestParsesRecordedWindowsResponses runs every message a real Windows server
// sent through this client's parser.
//
// This is the check with independent authority: those bytes were produced by
// Windows, not by this code, so unlike a round trip against this repository's
// own server it cannot pass by both halves sharing a mistake.
func TestParsesRecordedWindowsResponses(t *testing.T) {
	for _, name := range []string{"smb311-server2025.json", "smb302-server2012r2.json"} {
		t.Run(name, func(t *testing.T) {
			corpus := referenceCorpus(t, name)
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
					t.Errorf("%s response %d (%d bytes): header did not parse: %v", corpus.Peer, i, len(raw), err)
					continue
				}
				if !msg.Header.HasValidProtocolId() {
					t.Errorf("response %d does not carry the SMB2 protocol id: % x", i, raw[:4])
				}
				if got := int(msg.Header.StructureSize); got != 64 {
					t.Errorf("response %d header StructureSize = %d, want 64", i, got)
				}
			}
			t.Logf("parsed %d responses recorded from %s", len(responses), corpus.Peer)
		})
	}
}

// TestNegotiateRequestMatchesTheRecording compares the NEGOTIATE this client
// emits today against the one it emitted when the corpus was recorded.
//
// NEGOTIATE is the densest fingerprint in the protocol — the dialect list,
// capabilities, security mode and negotiate contexts all sit in one message,
// before any session exists — and it is almost entirely deterministic, so it can
// be pinned byte for byte with only the ClientGuid and the pre-auth salt excused.
func TestNegotiateRequestMatchesTheRecording(t *testing.T) {
	corpus := referenceCorpus(t, "smb311-server2025.json")
	recorded, err := corpus.Messages(wirediff.FromClient)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}
	want := recorded[0]

	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponse(t)}}
	c := newTestClient(ft)
	if err := c.Negotiate(); err != nil {
		t.Fatalf("Negotiate: %v", err)
	}
	got := c.lastSentBytes
	if len(got) == 0 {
		t.Fatal("the client recorded no sent bytes")
	}

	// The negotiate context list begins at NegotiateContextOffset, which is
	// header-relative; everything from there carries the fresh pre-auth salt.
	contextsOffset := int(binary.LittleEndian.Uint32(want[64+28 : 64+32]))
	mask := wirediff.SMB2NegotiateRequestMask(contextsOffset)

	if diffs := wirediff.Compare(got, want, mask); len(diffs) > 0 {
		t.Errorf("the NEGOTIATE this client emits differs from the recorded one.\n%s\nIf the change was intended, re-record the corpus (see testdata/reference/README.md).",
			wirediff.Report(got, want, diffs))
	}
}

// TestRecordedNegotiateShowsAnUnsetClientGuid documents a deviation the corpus
// makes visible: ClientGuid is emitted as sixteen zero bytes.
//
// A conforming client sends a stable, machine-scoped GUID; nothing in this
// repository ever assigns Client.ClientGuid, so the field goes out empty on
// every connection. The comparison above has to excuse the field to be useful at
// all, so this asserts the deviation separately rather than letting the mask
// quietly bury it. When the field is populated, this test is what will fail, and
// it should then be deleted.
func TestRecordedNegotiateShowsAnUnsetClientGuid(t *testing.T) {
	corpus := referenceCorpus(t, "smb311-server2025.json")
	recorded, err := corpus.Messages(wirediff.FromClient)
	if err != nil {
		t.Fatalf("reading corpus: %v", err)
	}

	guid := recorded[0][64+12 : 64+28]
	empty := true
	for _, b := range guid {
		if b != 0x00 {
			empty = false
			break
		}
	}
	if !empty {
		t.Errorf("ClientGuid is now % x; the known deviation is fixed and this test should be removed", guid)
	}
}
