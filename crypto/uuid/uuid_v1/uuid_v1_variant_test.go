package uuid_v1_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v1"
)

// RFC 9562 Appendix A.1 test vector: time 0x1EC9414C232AB00,
// clock sequence 0x33C8, node 9F6BDECED846.
func TestUUIDv1RFC9562Vector(t *testing.T) {
	const want = "c232ab00-9414-11ec-b3c8-9f6bdeced846"

	u := uuid_v1.UUIDv1{
		Time:     0x1EC9414C232AB00,
		ClockSeq: 0x33C8,
		NodeID:   [6]byte{0x9F, 0x6B, 0xDE, 0xCE, 0xD8, 0x46},
	}
	if got := u.String(); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}

	var parsed uuid_v1.UUIDv1
	if err := parsed.FromString(want); err != nil {
		t.Fatalf("FromString failed: %v", err)
	}
	if parsed.ClockSeq != 0x33C8 {
		t.Errorf("ClockSeq = 0x%04x, want 0x33c8", parsed.ClockSeq)
	}
	if parsed.Time != 0x1EC9414C232AB00 {
		t.Errorf("Time = 0x%x, want 0x1ec9414c232ab00", parsed.Time)
	}
}

func TestUUIDv1VariantAndFullClockSeq(t *testing.T) {
	for _, cs := range []uint16{0x0000, 0x0fff, 0x1000, 0x2abc, 0x3fff} {
		u := uuid_v1.UUIDv1{Time: 0x0cde89ab01234567, ClockSeq: cs}
		data, err := u.Marshal()
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		if data[8]&0xC0 != 0x80 {
			t.Errorf("clock seq 0x%04x: variant bits = %02b, want 10", cs, data[8]>>6)
		}
		var v uuid_v1.UUIDv1
		if _, err := v.Unmarshal(data); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if v.ClockSeq != cs {
			t.Errorf("clock seq round trip: got 0x%04x, want 0x%04x", v.ClockSeq, cs)
		}
	}
}
