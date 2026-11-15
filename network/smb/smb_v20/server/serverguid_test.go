package server

import "testing"

// TestServerGUIDIsValidV4 checks that an auto-generated ServerGUID is a
// well-formed RFC 4122 version 4 GUID. Raw random bytes (the previous behaviour)
// leave an arbitrary version/variant, which fingerprints the server: a real
// Windows ServerGuid always has version 4 and variant 10xx. The bytes are laid
// out as they travel on the wire, so byte 7 carries the version nibble and byte 8
// the variant bits.
func TestServerGUIDIsValidV4(t *testing.T) {
	s, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	g := s.config.ServerGUID
	if g == ([16]byte{}) {
		t.Fatal("ServerGUID was not generated")
	}

	if version := g[7] >> 4; version != 4 {
		t.Errorf("ServerGUID version nibble = %d, want 4 (bytes % x)", version, g)
	}
	if variant := g[8] >> 6; variant != 0b10 {
		t.Errorf("ServerGUID variant bits = %#b, want 10 (bytes % x)", variant, g)
	}
}
