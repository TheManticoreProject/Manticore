package nbdgm

import (
	"bytes"
	"net"
	"testing"
)

// TestReassemblyRejectsFragmentPastEnd checks that bytes received beyond the
// end fixed by the last fragment do not count towards completion: a datagram
// with a hole inside [0, total) must never be reported as reassembled.
func TestReassemblyRejectsFragmentPastEnd(t *testing.T) {
	ip := net.IPv4(10, 0, 0, 1)
	first := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst | FlagMore,
		DgmID:           7,
		SourceIP:        ip,
		SourceName:      Name{Name: "SRC"},
		DestinationName: Name{Name: "DST"},
	}
	// A middle fragment lying entirely past the eventual end of the datagram.
	beyond := &Datagram{
		MsgType:      MsgTypeDirectUnique,
		Flags:        FlagMore,
		DgmID:        7,
		SourceIP:     ip,
		PacketOffset: 100,
		UserData:     bytes.Repeat([]byte{'X'}, 50),
	}
	// The last fragment covers only [30, 40), leaving [0, 30) unfilled.
	last := &Datagram{
		MsgType:      MsgTypeDirectUnique,
		DgmID:        7,
		SourceIP:     ip,
		PacketOffset: 30,
		UserData:     bytes.Repeat([]byte{'Y'}, 10),
	}

	r := NewReassembler()
	for i, d := range []*Datagram{first, beyond} {
		if _, done, err := r.Add("10.0.0.1:138", d); err != nil || done {
			t.Fatalf("fragment %d: done=%v err=%v", i, done, err)
		}
	}
	a, done, err := r.Add("10.0.0.1:138", last)
	if done {
		t.Fatalf("datagram with a hole reported complete: USER_DATA = %q", a.UserData)
	}
	if err == nil {
		t.Error("expected an error for a fragment extending past the end of the datagram")
	}
}

// TestReassemblyRejectsFragmentAfterEnd checks that once the last fragment has
// fixed the total length, a later fragment extending past it is rejected
// rather than satisfying the byte count for the missing range.
func TestReassemblyRejectsFragmentAfterEnd(t *testing.T) {
	ip := net.IPv4(10, 0, 0, 1)
	first := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst | FlagMore,
		DgmID:           8,
		SourceIP:        ip,
		SourceName:      Name{Name: "SRC"},
		DestinationName: Name{Name: "DST"},
	}
	last := &Datagram{
		MsgType:      MsgTypeDirectUnique,
		DgmID:        8,
		SourceIP:     ip,
		PacketOffset: 30,
		UserData:     bytes.Repeat([]byte{'Y'}, 10),
	}
	beyond := &Datagram{
		MsgType:      MsgTypeDirectUnique,
		Flags:        FlagMore,
		DgmID:        8,
		SourceIP:     ip,
		PacketOffset: 100,
		UserData:     bytes.Repeat([]byte{'X'}, 50),
	}

	r := NewReassembler()
	for i, d := range []*Datagram{first, last} {
		if _, done, err := r.Add("10.0.0.1:138", d); err != nil || done {
			t.Fatalf("fragment %d: done=%v err=%v", i, done, err)
		}
	}
	a, done, err := r.Add("10.0.0.1:138", beyond)
	if done {
		t.Fatalf("datagram with a hole reported complete: USER_DATA = %q", a.UserData)
	}
	if err == nil {
		t.Error("expected an error for a fragment extending past the end of the datagram")
	}
}
