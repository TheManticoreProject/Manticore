package nbdgm

import (
	"net"
	"testing"
)

// TestFragmentRejectsOversizedUserData checks that USER_DATA longer than
// MaxUserDataSize is refused: PACKET_OFFSET is 16 bits wide, so fragments past
// offset 65535 would otherwise carry a wrapped offset.
func TestFragmentRejectsOversizedUserData(t *testing.T) {
	s := &Sender{
		SourceName: Name{Name: "SENDER"},
		SourceIP:   net.IPv4(10, 0, 0, 1),
		SourcePort: 138,
	}
	if _, err := s.Fragment(MsgTypeDirectUnique, Name{Name: "TARGET"}, make([]byte, MaxUserDataSize+1)); err == nil {
		t.Fatalf("Fragment accepted %d bytes of USER_DATA (> MaxUserDataSize %d)", MaxUserDataSize+1, MaxUserDataSize)
	}
}

// TestFragmentAcceptsMaxUserData checks that exactly MaxUserDataSize bytes of
// USER_DATA are still fragmented and reassemble to the original payload.
func TestFragmentAcceptsMaxUserData(t *testing.T) {
	s := &Sender{
		SourceName: Name{Name: "SENDER"},
		SourceIP:   net.IPv4(10, 0, 0, 1),
		SourcePort: 138,
	}
	userData := make([]byte, MaxUserDataSize)
	for i := range userData {
		userData[i] = byte(i)
	}
	frags, err := s.Fragment(MsgTypeDirectUnique, Name{Name: "TARGET"}, userData)
	if err != nil {
		t.Fatalf("Fragment: %v", err)
	}

	r := NewReassembler()
	var assembled *Datagram
	for i, frag := range frags {
		var d Datagram
		if _, err := d.Unmarshal(frag); err != nil {
			t.Fatalf("Unmarshal fragment %d: %v", i, err)
		}
		a, done, err := r.Add("10.0.0.1:138", &d)
		if err != nil {
			t.Fatalf("Add fragment %d: %v", i, err)
		}
		if done {
			assembled = a
		}
	}
	if assembled == nil || len(assembled.UserData) != MaxUserDataSize {
		t.Fatal("maximum-size USER_DATA did not reassemble")
	}
	for i, b := range assembled.UserData {
		if b != byte(i) {
			t.Fatalf("reassembled USER_DATA differs at offset %d", i)
		}
	}
}
