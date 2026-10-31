package nbns

import (
	"net"
	"testing"
	"time"
)

// TestUDPServerAnswersEachDatagramIndependently sends a burst of NAME QUERY
// REQUESTs alternating between two registered names, each with its own
// transaction ID, and confirms every response answers the query that carried its
// NAME_TRN_ID. The serve loop reuses one read buffer, so a handler that parsed
// that buffer in place would see it overwritten by the next datagram and answer
// a different query under the original transaction ID.
func TestUDPServerAnswersEachDatagramIndependently(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	if err := nbns.RegisterName("AAAA", "", Unique, net.ParseIP("10.0.0.1"), time.Hour); err != nil {
		t.Fatalf("RegisterName AAAA: %v", err)
	}
	if err := nbns.RegisterName("BBBB", "", Unique, net.ParseIP("10.0.0.2"), time.Hour); err != nil {
		t.Fatalf("RegisterName BBBB: %v", err)
	}

	udp, err := NewUDPServer("127.0.0.1:0", nbns)
	if err != nil {
		t.Fatalf("NewUDPServer: %v", err)
	}
	if err := udp.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer udp.Stop()

	conn, err := net.DialUDP("udp", nil, udp.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("DialUDP: %v", err)
	}
	defer conn.Close()

	const queries = 2000
	client := NewClient()
	want := make(map[uint16]byte, queries)
	for i := 0; i < queries; i++ {
		name, lastOctet := "AAAA", byte(1)
		if i%2 == 1 {
			name, lastOctet = "BBBB", 2
		}
		req, err := client.BuildNameQueryRequest(name, 0x20, "")
		if err != nil {
			t.Fatalf("BuildNameQueryRequest: %v", err)
		}
		req.Header.TransactionID = uint16(i)
		want[uint16(i)] = lastOctet
		data, err := req.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if _, err := conn.Write(data); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	received, mismatched := 0, 0
	buf := make([]byte, 1024)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			t.Fatalf("SetReadDeadline: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		var resp NBNSPacket
		if _, err := resp.Unmarshal(buf[:n]); err != nil || len(resp.Answers) != 1 {
			mismatched++
			continue
		}
		received++
		ip, err := ParseIPFromRData(resp.Answers[0].RData)
		if err != nil || ip.To4()[3] != want[resp.Header.TransactionID] {
			mismatched++
		}
	}

	if received == 0 {
		t.Fatal("no responses received")
	}
	if mismatched != 0 {
		t.Fatalf("%d of %d responses answered a different query than their NAME_TRN_ID", mismatched, received)
	}
}
