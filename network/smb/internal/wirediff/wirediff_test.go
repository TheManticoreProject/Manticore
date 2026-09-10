package wirediff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareIgnoresOnlyMaskedBytes(t *testing.T) {
	want := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	got := []byte{0x01, 0xFF, 0x03, 0xFF, 0x05}

	// With no mask both differing bytes are reported.
	diffs := Compare(got, want, nil)
	if len(diffs) != 2 {
		t.Fatalf("unmasked comparison found %d differences, want 2", len(diffs))
	}
	if diffs[0].Offset != 1 || diffs[0].Got != 0xFF || diffs[0].Want != 0x02 {
		t.Errorf("first difference = %+v, want offset 1 got ff want 02", diffs[0])
	}

	// Masking one span excuses exactly that byte.
	diffs = Compare(got, want, Mask{{Name: "volatile", Start: 1, Len: 1}})
	if len(diffs) != 1 || diffs[0].Offset != 3 {
		t.Fatalf("masked comparison = %+v, want only offset 3", diffs)
	}
}

func TestCompareReportsLengthMismatch(t *testing.T) {
	if diffs := Compare([]byte{1, 2, 3}, []byte{1, 2}, nil); len(diffs) != 1 || diffs[0].Offset != 2 {
		t.Errorf("a longer emission reported %+v, want one difference at offset 2", diffs)
	}
	if diffs := Compare([]byte{1, 2}, []byte{1, 2, 3}, nil); len(diffs) != 1 || diffs[0].Offset != 2 {
		t.Errorf("a shorter emission reported %+v, want one difference at offset 2", diffs)
	}
	// A trailing span may be masked, which is how a variable-length tail is excused.
	if diffs := Compare([]byte{1, 2, 3}, []byte{1, 2}, Mask{{Name: "tail", Start: 2, Len: 64}}); len(diffs) != 0 {
		t.Errorf("masked tail reported %+v, want none", diffs)
	}
}

func TestMaskCovers(t *testing.T) {
	mask := Mask{{Name: "signature", Start: 48, Len: 16}}
	if name, ok := mask.Covers(48); !ok || name != "signature" {
		t.Errorf("start of span: got (%q,%v)", name, ok)
	}
	if name, ok := mask.Covers(63); !ok || name != "signature" {
		t.Errorf("end of span: got (%q,%v)", name, ok)
	}
	if _, ok := mask.Covers(64); ok {
		t.Error("offset past the span reported as covered")
	}
	if _, ok := mask.Covers(47); ok {
		t.Error("offset before the span reported as covered")
	}
}

// TestSMB2HeaderMaskPinsTheFingerprintFields is the guard on the mask itself: a
// mask that excused CreditCharge, Credit or Flags would silently defeat the
// comparison, since those are exactly the fields that distinguish one
// implementation from another.
func TestSMB2HeaderMaskPinsTheFingerprintFields(t *testing.T) {
	mask := SMB2HeaderMask()

	pinned := map[string]int{
		"CreditCharge": 6,
		"Command":      12,
		"Credit":       14,
		"Flags":        16,
		"NextCommand":  20,
	}
	for name, offset := range pinned {
		if masked, ok := mask.Covers(offset); ok {
			t.Errorf("%s at offset %d is masked as %q, so a change in it would go unnoticed", name, offset, masked)
		}
	}

	for name, offset := range map[string]int{"MessageId": 24, "TreeId": 36, "SessionId": 40, "Signature": 48} {
		if _, ok := mask.Covers(offset); !ok {
			t.Errorf("%s at offset %d is not masked, so every run would differ", name, offset)
		}
	}
}

func TestSMB2NegotiateRequestMaskKeepsCapabilitiesPinned(t *testing.T) {
	mask := SMB2NegotiateRequestMask(0)

	// ClientGuid is excused, the fields around it are not.
	if _, ok := mask.Covers(smb2HeaderSize + 12); !ok {
		t.Error("ClientGuid is not masked")
	}
	for name, offset := range map[string]int{
		"StructureSize": smb2HeaderSize + 0,
		"DialectCount":  smb2HeaderSize + 2,
		"SecurityMode":  smb2HeaderSize + 4,
		"Capabilities":  smb2HeaderSize + 8,
	} {
		if masked, ok := mask.Covers(offset); ok {
			t.Errorf("NEGOTIATE %s is masked as %q; it is a fingerprint field and must be pinned", name, masked)
		}
	}
}

func TestLoadCorpus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corpus.json")
	if err := os.WriteFile(path, []byte(`{
		"name": "example",
		"peer": "a peer",
		"frames": [
			{"direction": "client", "message": "fe534d4200"},
			{"direction": "server", "message": "fe534d4201"}
		]
	}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	corpus, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if corpus.Name != "example" {
		t.Errorf("Name = %q", corpus.Name)
	}

	client, err := corpus.Messages(FromClient)
	if err != nil {
		t.Fatalf("Messages: %v", err)
	}
	if len(client) != 1 || client[0][4] != 0x00 {
		t.Errorf("client messages = %x", client)
	}
	server, err := corpus.Messages(FromServer)
	if err != nil {
		t.Fatalf("Messages: %v", err)
	}
	if len(server) != 1 || server[0][4] != 0x01 {
		t.Errorf("server messages = %x", server)
	}
}

func TestLoadRejectsAnEmptyCorpus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(`{"name":"x","frames":[]}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Error("an empty corpus loaded without error")
	}
}

func TestReportNamesTheOffsets(t *testing.T) {
	got, want := []byte{1, 2, 3}, []byte{1, 9, 3}
	report := Report(got, want, Compare(got, want, nil))
	if !strings.Contains(report, "offset 1") || !strings.Contains(report, "got 02") || !strings.Contains(report, "want 09") {
		t.Errorf("report does not identify the difference:\n%s", report)
	}
	if Report(got, got, nil) != "" {
		t.Error("an identical pair produced a report")
	}
}
