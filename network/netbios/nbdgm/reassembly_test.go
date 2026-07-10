package nbdgm

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// buildFragments produces the wire fragments for a DIRECT_UNIQUE datagram whose
// USER_DATA is split into three pieces by a small MaxFragmentSize, exercising
// the FIRST/MORE + PACKET_OFFSET path.
func buildFragments(t *testing.T, userData []byte) ([][]byte, *Sender, Name) {
	t.Helper()
	s := &Sender{
		SourceName:      Name{Name: "SENDER", Suffix: 0x00},
		SourceIP:        net.IPv4(10, 7, 0, 20),
		SourcePort:      138,
		NodeType:        NodeTypeB,
		MaxFragmentSize: 128, // small enough to force multiple fragments
	}
	dst := Name{Name: "TARGET", Suffix: 0x20}
	frags, err := s.Fragment(MsgTypeDirectUnique, dst, userData)
	if err != nil {
		t.Fatalf("Fragment: %v", err)
	}
	if len(frags) < 3 {
		t.Fatalf("expected at least 3 fragments, got %d", len(frags))
	}
	return frags, s, dst
}

func TestReassemblyThreeFragments(t *testing.T) {
	// A payload larger than a single 128-byte fragment can hold.
	userData := bytes.Repeat([]byte("ABCDEFGHIJ"), 30) // 300 bytes
	frags, _, _ := buildFragments(t, userData)

	// Verify the fragment flags: FIRST only on the first, MORE on all but last.
	for i, frag := range frags {
		var d Datagram
		if _, err := d.Unmarshal(frag); err != nil {
			t.Fatalf("Unmarshal fragment %d: %v", i, err)
		}
		wantFirst := i == 0
		if d.IsFirst() != wantFirst {
			t.Errorf("fragment %d FIRST = %v, want %v", i, d.IsFirst(), wantFirst)
		}
		wantMore := i != len(frags)-1
		if d.HasMore() != wantMore {
			t.Errorf("fragment %d MORE = %v, want %v", i, d.HasMore(), wantMore)
		}
	}

	r := NewReassembler()
	src := "10.7.0.20:138"
	var assembled *Datagram
	for i, frag := range frags {
		var d Datagram
		if _, err := d.Unmarshal(frag); err != nil {
			t.Fatalf("Unmarshal fragment %d: %v", i, err)
		}
		a, done, err := r.Add(src, &d)
		if err != nil {
			t.Fatalf("Add fragment %d: %v", i, err)
		}
		if done != (i == len(frags)-1) {
			t.Fatalf("fragment %d done = %v, want %v", i, done, i == len(frags)-1)
		}
		if done {
			assembled = a
		}
	}

	if assembled == nil {
		t.Fatal("reassembly never completed")
	}
	if !bytes.Equal(assembled.UserData, userData) {
		t.Errorf("reassembled USER_DATA mismatch: got %d bytes, want %d", len(assembled.UserData), len(userData))
	}
	if assembled.SourceName.Name != "SENDER" || assembled.DestinationName.Name != "TARGET" {
		t.Errorf("reassembled names = src %+v dst %+v", assembled.SourceName, assembled.DestinationName)
	}
	if assembled.DestinationName.Suffix != 0x20 {
		t.Errorf("reassembled DESTINATION_NAME suffix = 0x%02x, want 0x20", assembled.DestinationName.Suffix)
	}
}

func TestReassemblyOutOfOrder(t *testing.T) {
	userData := bytes.Repeat([]byte("0123456789"), 30)
	frags, _, _ := buildFragments(t, userData)

	// Feed the fragments in reverse order; reassembly by PACKET_OFFSET must
	// still reconstruct the original payload.
	r := NewReassembler()
	src := "10.7.0.20:138"
	var assembled *Datagram
	for i := len(frags) - 1; i >= 0; i-- {
		var d Datagram
		if _, err := d.Unmarshal(frags[i]); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		a, done, err := r.Add(src, &d)
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if done {
			assembled = a
		}
	}
	if assembled == nil {
		t.Fatal("out-of-order reassembly never completed")
	}
	if !bytes.Equal(assembled.UserData, userData) {
		t.Error("out-of-order reassembled USER_DATA mismatch")
	}
}

func TestReassemblyUnfragmentedFastPath(t *testing.T) {
	d := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst, // FIRST set, MORE clear => complete in one packet
		SourceIP:        net.IPv4(10, 0, 0, 1),
		SourceName:      Name{Name: "A"},
		DestinationName: Name{Name: "B"},
		UserData:        []byte("single"),
	}
	r := NewReassembler()
	a, done, err := r.Add("10.0.0.1:138", d)
	if err != nil || !done {
		t.Fatalf("fast path: done=%v err=%v", done, err)
	}
	if !bytes.Equal(a.UserData, []byte("single")) {
		t.Errorf("USER_DATA = %q", a.UserData)
	}
}

func TestReassemblyTimeoutEviction(t *testing.T) {
	r := NewReassembler()
	r.SetTimeout(1 * time.Millisecond)

	// A first fragment that expects more, left incomplete.
	d := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst | FlagMore,
		DgmID:           0x1234,
		SourceIP:        net.IPv4(10, 0, 0, 2),
		SourceName:      Name{Name: "A"},
		DestinationName: Name{Name: "B"},
		UserData:        []byte("part"),
	}
	if _, done, err := r.Add("10.0.0.2:138", d); err != nil || done {
		t.Fatalf("first fragment: done=%v err=%v", done, err)
	}

	time.Sleep(5 * time.Millisecond)

	// Any subsequent Add triggers eviction of the expired entry; a new first
	// fragment for the same key starts fresh (no stale bytes leak in).
	r.mu.Lock()
	before := len(r.pending)
	r.mu.Unlock()
	if before != 1 {
		t.Fatalf("expected 1 pending entry, got %d", before)
	}

	// Trigger eviction via a fragment for a different key.
	other := &Datagram{
		MsgType:         MsgTypeDirectUnique,
		Flags:           FlagFirst | FlagMore,
		DgmID:           0x9999,
		SourceIP:        net.IPv4(10, 0, 0, 3),
		SourceName:      Name{Name: "C"},
		DestinationName: Name{Name: "D"},
		UserData:        []byte("x"),
	}
	if _, _, err := r.Add("10.0.0.3:138", other); err != nil {
		t.Fatalf("Add other: %v", err)
	}
	r.mu.Lock()
	_, expiredStillThere := r.pending[reassemblyKey{source: "10.0.0.2:138", dgmID: 0x1234}]
	r.mu.Unlock()
	if expiredStillThere {
		t.Error("expired reassembly entry was not evicted")
	}
}
