package server

import "testing"

// TestStableServerGUIDIsStableAndValid checks that the host-derived ServerGUID is
// the same on repeated calls (so it survives a process restart) and is a
// well-formed v4 GUID. Before this, each start generated a fresh GUID, which an
// observer watching one host over time could see change.
func TestStableServerGUIDIsStableAndValid(t *testing.T) {
	a := StableServerGUID()
	b := StableServerGUID()
	if !a.Equal(&b) {
		t.Fatalf("ServerGUID changed between calls: %s vs %s", a.ToFormatN(), b.ToFormatN())
	}

	bytes := a.ToBytes()
	if version := bytes[7] >> 4; version != 4 {
		t.Errorf("version nibble = %d, want 4", version)
	}
	if variant := bytes[8] >> 6; variant != 0b10 {
		t.Errorf("variant bits = %#b, want 10", variant)
	}
}

// TestNewServerServerGUIDStableAcrossRestarts checks two independently
// constructed servers on the same host report the same ServerGUID, standing in
// for a restart.
func TestNewServerServerGUIDStableAcrossRestarts(t *testing.T) {
	s1, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	s2, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if s1.config.ServerGUID != s2.config.ServerGUID {
		t.Fatalf("ServerGUID differs across restarts: %s vs %s",
			s1.config.ServerGUID.ToFormatN(), s2.config.ServerGUID.ToFormatN())
	}
}
