package nbns

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"
)

// TestUDPServerTruncatedResponseSetsTC registers enough names that the NODE
// STATUS RESPONSE exceeds MaxUDPSize and confirms the datagram the server sends
// is cut to MaxUDPSize with the TC (truncated) bit set in the header on the wire
// (RFC 1002 4.2.1.1), so the requestor knows to retry over TCP.
func TestUDPServerTruncatedResponseSetsTC(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	for i := 0; i < 50; i++ {
		if err := nbns.RegisterName(fmt.Sprintf("NAME%02d", i), "", Unique, net.ParseIP("10.0.0.1"), time.Hour); err != nil {
			t.Fatalf("RegisterName: %v", err)
		}
	}

	udp, err := NewUDPServer("127.0.0.1:0", nbns)
	if err != nil {
		t.Fatalf("NewUDPServer: %v", err)
	}
	udp.EnableNodeStatus(net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0x00, 0x01})
	if err := udp.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer udp.Stop()

	conn, err := net.DialUDP("udp", nil, udp.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("DialUDP: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write(buildNodeStatusRequest(0x1234)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if n != MaxUDPSize {
		t.Fatalf("response length = %d, want it truncated to %d", n, MaxUDPSize)
	}
	flags := binary.BigEndian.Uint16(buf[2:4])
	if flags&FlagTruncated == 0 {
		t.Errorf("truncated response flags = 0x%04x, want TC (0x%04x) set", flags, FlagTruncated)
	}
	if flags&FlagResponse == 0 {
		t.Errorf("truncated response flags = 0x%04x, want R (0x%04x) set", flags, FlagResponse)
	}
}
