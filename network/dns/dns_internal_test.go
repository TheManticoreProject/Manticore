package dns

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestDNSServerAddress(t *testing.T) {
	tests := []struct {
		server string
		want   string
	}{
		{"8.8.8.8", "8.8.8.8:53"},
		{"8.8.8.8:5353", "8.8.8.8:5353"},
		{"::1", "[::1]:53"},
		{"[::1]:5353", "[::1]:5353"},
		{"fe80::1", "[fe80::1]:53"},
		{"dns.example.com", "dns.example.com:53"},
		{"[::1]", "[::1]:53"},
		{"fe80::1%eth0", "[fe80::1%eth0]:53"},
		{"", ":53"},
	}
	for _, tt := range tests {
		if got := dnsServerAddress(tt.server); got != tt.want {
			t.Errorf("dnsServerAddress(%q) = %q, want %q", tt.server, got, tt.want)
		}
	}
}

// serveStubDNS answers every A query on conn with 192.0.2.53 and every
// other query type with an empty NOERROR response.
func serveStubDNS(conn net.PacketConn) {
	buf := make([]byte, 512)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}
		if n < 12 {
			continue
		}
		// Locate the end of the question section (QNAME + QTYPE + QCLASS).
		off := 12
		for off < n && buf[off] != 0 {
			off += int(buf[off]) + 1
		}
		off += 5
		if off > n {
			continue
		}
		qtype := binary.BigEndian.Uint16(buf[off-4 : off-2])

		resp := make([]byte, 0, off+16)
		resp = append(resp, buf[0:2]...) // ID
		resp = append(resp, 0x81, 0x80)  // QR, RD, RA, NOERROR
		resp = append(resp, 0x00, 0x01)  // QDCOUNT
		if qtype == 1 {
			resp = append(resp, 0x00, 0x01) // ANCOUNT
		} else {
			resp = append(resp, 0x00, 0x00)
		}
		resp = append(resp, 0x00, 0x00, 0x00, 0x00) // NSCOUNT, ARCOUNT
		resp = append(resp, buf[12:off]...)         // Question
		if qtype == 1 {
			resp = append(resp,
				0xc0, 0x0c, // Name pointer to question
				0x00, 0x01, // TYPE A
				0x00, 0x01, // CLASS IN
				0x00, 0x00, 0x00, 0x3c, // TTL
				0x00, 0x04, // RDLENGTH
				192, 0, 2, 53,
			)
		}
		conn.WriteTo(resp, addr)
	}
}

// An IPv6 server given with an explicit port still resolves end to end. The bare
// "::1" form needs port 53 and is covered by TestDNSServerAddress instead.
func TestDNSLookupIPv6ServerWithPort(t *testing.T) {
	conn, err := net.ListenPacket("udp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 loopback not available: %v", err)
	}
	defer conn.Close()
	go serveStubDNS(conn)

	server := conn.LocalAddr().String() // "[::1]:<port>"
	ips := DNSLookup("manticore-stub.example.com", server)
	if len(ips) != 1 || ips[0] != "192.0.2.53" {
		t.Fatalf("DNSLookup via %s = %v, want [192.0.2.53]", server, ips)
	}
}
