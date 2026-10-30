package uuid_v2_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v2"
)

// DCE 1.1 Security: byte 8 is clock_seq_hi_and_reserved, carrying the
// variant (0b10) and a 6-bit clock sequence; byte 9 is the local domain.
func TestUUIDv2VariantAndFullClock(t *testing.T) {
	for _, clock := range []uint8{0x00, 0x0f, 0x10, 0x2a, 0x3f} {
		u := uuid_v2.UUIDv2{LocalDomainNumber: 1000, Time: 0x0cde89ab00000000, Clock: clock, LocalDomain: 1}
		data, err := u.Marshal()
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}
		if want := 0x80 | clock; data[8] != want {
			t.Errorf("clock 0x%02x: byte 8 = 0x%02x, want 0x%02x", clock, data[8], want)
		}
		if data[9] != 1 {
			t.Errorf("clock 0x%02x: local domain byte = 0x%02x, want 0x01", clock, data[9])
		}
		var v uuid_v2.UUIDv2
		if _, err := v.Unmarshal(data); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if v.Clock != clock {
			t.Errorf("clock round trip: got 0x%02x, want 0x%02x", v.Clock, clock)
		}
	}
}
