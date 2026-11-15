package guid

import "testing"

// TestNewGUIDFromSeedIsDeterministic checks the same seed always yields the same
// GUID (so a server GUID derived from the host identity is stable across
// restarts) and that different seeds yield different GUIDs.
func TestNewGUIDFromSeedIsDeterministic(t *testing.T) {
	a := NewGUIDFromSeed([]byte("host=alpha|00:11:22:33:44:55"))
	b := NewGUIDFromSeed([]byte("host=alpha|00:11:22:33:44:55"))
	if !a.Equal(b) {
		t.Fatalf("same seed produced different GUIDs: %s vs %s", a.ToFormatN(), b.ToFormatN())
	}

	c := NewGUIDFromSeed([]byte("host=beta|00:11:22:33:44:55"))
	if a.Equal(c) {
		t.Fatalf("different seeds produced the same GUID: %s", a.ToFormatN())
	}
}

// TestNewGUIDFromSeedIsValidV4 checks the derived GUID carries the RFC 4122
// version 4 and variant bits, like NewGUID.
func TestNewGUIDFromSeedIsValidV4(t *testing.T) {
	g := NewGUIDFromSeed([]byte("anything"))
	b := g.ToBytes()
	if version := b[7] >> 4; version != 4 {
		t.Errorf("version nibble = %d, want 4 (bytes % x)", version, b)
	}
	if variant := b[8] >> 6; variant != 0b10 {
		t.Errorf("variant bits = %#b, want 10 (bytes % x)", variant, b)
	}
}
