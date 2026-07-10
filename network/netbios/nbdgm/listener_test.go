package nbdgm

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// TestSenderListenerRoundTrip proves the datagram transport end to end over
// real UDP sockets: a Listener on an ephemeral loopback port receives a
// DIRECT_UNIQUE datagram sent by a Sender, and the exact USER_DATA and decoded
// SOURCE_NAME/DESTINATION_NAME are recovered.
func TestSenderListenerRoundTrip(t *testing.T) {
	received := make(chan *Datagram, 1)
	l, err := NewListener("127.0.0.1:0", func(_ *net.UDPAddr, d *Datagram) {
		received <- d
	})
	if err != nil {
		t.Fatalf("NewListener: %v", err)
	}
	if err := l.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer l.Stop()

	dest := l.LocalAddr().String()

	s := &Sender{
		SourceName: Name{Name: "SENDER", Suffix: 0x00},
		SourceIP:   net.IPv4(127, 0, 0, 1),
		SourcePort: 138,
		NodeType:   NodeTypeB,
	}
	payload := []byte("the quick brown fox")
	if err := s.SendDirectUnique(dest, Name{Name: "TARGET", Suffix: 0x20}, payload); err != nil {
		t.Fatalf("SendDirectUnique: %v", err)
	}

	select {
	case d := <-received:
		if !bytes.Equal(d.UserData, payload) {
			t.Errorf("USER_DATA = %q, want %q", d.UserData, payload)
		}
		if d.SourceName.Name != "SENDER" {
			t.Errorf("SOURCE_NAME = %q, want SENDER", d.SourceName.Name)
		}
		if d.DestinationName.Name != "TARGET" || d.DestinationName.Suffix != 0x20 {
			t.Errorf("DESTINATION_NAME = %+v", d.DestinationName)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for datagram")
	}
}

// TestSenderListenerFragmentedRoundTrip proves fragmentation end to end: a
// USER_DATA payload larger than a single UDP datagram is fragmented by the
// Sender, delivered over loopback, and reassembled byte-for-byte by the
// Listener before the callback fires exactly once.
func TestSenderListenerFragmentedRoundTrip(t *testing.T) {
	received := make(chan *Datagram, 4)
	l, err := NewListener("127.0.0.1:0", func(_ *net.UDPAddr, d *Datagram) {
		received <- d
	})
	if err != nil {
		t.Fatalf("NewListener: %v", err)
	}
	if err := l.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer l.Stop()

	dest := l.LocalAddr().String()

	s := &Sender{
		SourceName:      Name{Name: "SENDER", Suffix: 0x00},
		SourceIP:        net.IPv4(127, 0, 0, 1),
		SourcePort:      138,
		NodeType:        NodeTypeM,
		MaxFragmentSize: 200, // force several fragments
	}

	// A distinctive, position-dependent payload so a reordering or offset bug
	// would corrupt the reassembly.
	var payload []byte
	for i := 0; i < 1000; i++ {
		payload = append(payload, byte(i%251))
	}

	dst := Name{Name: "TARGET", Suffix: 0x20}
	frags, err := s.Fragment(MsgTypeDirectUnique, dst, payload)
	if err != nil {
		t.Fatalf("Fragment: %v", err)
	}
	if len(frags) < 4 {
		t.Fatalf("expected the payload to fragment into >=4 packets, got %d", len(frags))
	}

	if err := s.SendDirectUnique(dest, dst, payload); err != nil {
		t.Fatalf("SendDirectUnique: %v", err)
	}

	select {
	case d := <-received:
		if !bytes.Equal(d.UserData, payload) {
			t.Errorf("reassembled USER_DATA mismatch: got %d bytes, want %d", len(d.UserData), len(payload))
		}
		if d.NodeType() != NodeTypeM {
			t.Errorf("NodeType = %d, want %d", d.NodeType(), NodeTypeM)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for reassembled datagram")
	}

	// The callback must fire exactly once for one logical datagram, even though
	// it arrived as several UDP packets.
	select {
	case <-received:
		t.Fatal("callback fired more than once for a single datagram")
	case <-time.After(150 * time.Millisecond):
	}
}

// TestListenerDeliversQueryRequest confirms a non-fragmenting datagram type
// (DATAGRAM QUERY REQUEST) is delivered as a single packet with its
// DESTINATION_NAME decoded.
func TestListenerDeliversQueryRequest(t *testing.T) {
	received := make(chan *Datagram, 1)
	l, err := NewListener("127.0.0.1:0", func(_ *net.UDPAddr, d *Datagram) {
		received <- d
	})
	if err != nil {
		t.Fatalf("NewListener: %v", err)
	}
	if err := l.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer l.Stop()

	d := &Datagram{
		MsgType:         MsgTypeQueryRequest,
		Flags:           FlagFirst,
		DgmID:           0x4242,
		SourceIP:        net.IPv4(127, 0, 0, 1),
		SourcePort:      138,
		DestinationName: Name{Name: "WORKGROUP", Suffix: 0x1D},
	}
	wire, err := d.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, l.LocalAddr())
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()
	if _, err := conn.Write(wire); err != nil {
		t.Fatalf("Write: %v", err)
	}

	select {
	case got := <-received:
		if got.MsgType != MsgTypeQueryRequest {
			t.Errorf("MSG_TYPE = 0x%02x", got.MsgType)
		}
		if got.DestinationName.Name != "WORKGROUP" || got.DestinationName.Suffix != 0x1D {
			t.Errorf("DESTINATION_NAME = %+v", got.DestinationName)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for query request")
	}
}
