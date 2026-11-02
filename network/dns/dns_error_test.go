package dns_test

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dns"
)

// serveDNS answers A queries for "ok.test." with 10.1.2.3 and every other query
// with NXDOMAIN, on a local UDP port, until the test ends.
func serveDNS(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 512)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			q := buf[:n]
			if len(q) < 12 {
				continue
			}
			// Question name ends at the first zero label; then QTYPE and QCLASS.
			end := 12
			for end < len(q) && q[end] != 0 {
				end += int(q[end]) + 1
			}
			if end+5 > len(q) {
				continue
			}
			question := q[12 : end+5]
			qtype := binary.BigEndian.Uint16(q[end+1:])
			name := strings.ToLower(string(q[13:end]))
			resp := append([]byte(nil), q[:2]...)
			if strings.HasPrefix(name, "ok\x04test") && qtype == 1 {
				resp = append(resp, 0x81, 0x80, 0, 1, 0, 1, 0, 0, 0, 0)
				resp = append(resp, question...)
				resp = append(resp, 0xc0, 0x0c, 0, 1, 0, 1, 0, 0, 0, 60, 0, 4, 10, 1, 2, 3)
			} else if strings.HasPrefix(name, "ok\x04test") {
				resp = append(resp, 0x81, 0x80, 0, 1, 0, 0, 0, 0, 0, 0) // no AAAA records
				resp = append(resp, question...)
			} else {
				resp = append(resp, 0x81, 0x83, 0, 1, 0, 0, 0, 0, 0, 0) // NXDOMAIN
				resp = append(resp, question...)
			}
			pc.WriteTo(resp, addr)
		}
	}()
	return pc.LocalAddr().String()
}

// TestDNSLookupWithErrorReportsFailures checks a resolved name returns its
// addresses with no error, and that NXDOMAIN and an unreachable server are
// reported as errors instead of an empty, error-free result.
func TestDNSLookupWithErrorReportsFailures(t *testing.T) {
	server := serveDNS(t)

	ips, err := dns.DNSLookupWithError("ok.test", server)
	if err != nil || len(ips) != 1 || ips[0] != "10.1.2.3" {
		t.Fatalf("ok.test: ips=%v err=%v, want [10.1.2.3] and no error", ips, err)
	}
	if got := dns.DNSLookup("ok.test", server); len(got) != 1 || got[0] != "10.1.2.3" {
		t.Errorf("DNSLookup(ok.test) = %v", got)
	}

	ips, err = dns.DNSLookupWithError("missing.test", server)
	var dnsErr *net.DNSError
	if err == nil || !errors.As(err, &dnsErr) || !dnsErr.IsNotFound {
		t.Errorf("missing.test: ips=%v err=%v, want a not-found *net.DNSError", ips, err)
	}
	if got := dns.DNSLookup("missing.test", server); len(got) != 0 {
		t.Errorf("DNSLookup(missing.test) = %v, want none", got)
	}

	// No server listens on the port: the lookup fails immediately with
	// "connection refused" instead of returning an empty, error-free answer.
	closed, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closedAddr := closed.LocalAddr().String()
	closed.Close()
	if ips, err := dns.DNSLookupWithError("ok.test", closedAddr); err == nil {
		t.Errorf("closed port: ips=%v with no error", ips)
	}
}
