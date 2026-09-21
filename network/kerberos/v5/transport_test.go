package kerberos

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// mustMarshalKRBError builds a wire KRB-ERROR with the given error code.
func mustMarshalKRBError(t *testing.T, code int) []byte {
	t.Helper()
	e := &messages.KRBError{
		ErrorCode: code,
		STime:     time.Now().UTC(),
		Realm:     "CORP.LOCAL",
		SName:     srvName("krbtgt", "CORP.LOCAL"),
	}
	b, err := e.Marshal()
	if err != nil {
		t.Fatalf("marshal KRB-ERROR: %v", err)
	}
	return b
}

// TestShouldRetryOverTCP covers the UDP->TCP decision: transport failure, empty
// datagram, and KRB_ERR_RESPONSE_TOO_BIG force a TCP retry; a complete protocol
// error or normal AS-REP does not.
func TestShouldRetryOverTCP(t *testing.T) {
	tooBig := mustMarshalKRBError(t, messages.ErrResponseTooBig)
	preauthRequired := mustMarshalKRBError(t, messages.ErrPreauthRequired)

	tests := []struct {
		name string
		resp []byte
		err  error
		want bool
	}{
		{"udp error", nil, errors.New("timeout"), true},
		{"empty datagram", []byte{}, nil, true},
		{"response too big", tooBig, nil, true},
		{"complete krb-error reply", preauthRequired, nil, false},
		{"as-rep reply", []byte{0x6b, 0x01, 0x02}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetryOverTCP(tt.resp, tt.err); got != tt.want {
				t.Errorf("shouldRetryOverTCP = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestKDCSendAddrReturnsUDPProtocolError verifies that a complete KRB-ERROR
// received over UDP is not discarded in favour of a failing TCP retry. The
// loopback port deliberately has a UDP listener but no TCP listener.
func TestKDCSendAddrReturnsUDPProtocolError(t *testing.T) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	want := mustMarshalKRBError(t, messages.ErrPreauthRequired)
	serverErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 1024)
		_, peer, err := conn.ReadFromUDP(buf)
		if err == nil {
			_, err = conn.WriteToUDP(want, peer)
		}
		serverErr <- err
	}()

	got, err := kdcSendAddr("127.0.0.1", conn.LocalAddr().(*net.UDPAddr).Port, []byte{0x01})
	if err != nil {
		t.Fatalf("kdcSendAddr discarded the UDP response: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("UDP test server: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("kdcSendAddr returned %x, want UDP KRB-ERROR %x", got, want)
	}
}

// TestResponseTooBigOverUDP verifies KRB_ERR_RESPONSE_TOO_BIG is recognised while
// other KRB-ERRORs and non-error replies are not.
func TestResponseTooBigOverUDP(t *testing.T) {
	if !responseTooBigOverUDP(mustMarshalKRBError(t, messages.ErrResponseTooBig)) {
		t.Error("expected RESPONSE_TOO_BIG to be recognised")
	}
	if responseTooBigOverUDP(mustMarshalKRBError(t, messages.ErrPreauthRequired)) {
		t.Error("PREAUTH_REQUIRED must not be classed as RESPONSE_TOO_BIG")
	}
	if responseTooBigOverUDP([]byte{0x6b, 0x01, 0x02}) {
		t.Error("a non-KRB-ERROR reply must not be classed as RESPONSE_TOO_BIG")
	}
	if responseTooBigOverUDP(nil) {
		t.Error("empty reply must not be classed as RESPONSE_TOO_BIG")
	}
}

// TestKRBErrorCode confirms the error code is extracted from a KRB-ERROR and that
// non-error bytes report ok=false.
func TestKRBErrorCode(t *testing.T) {
	code, ok := krbErrorCode(mustMarshalKRBError(t, messages.ErrSkew))
	if !ok || code != messages.ErrSkew {
		t.Errorf("krbErrorCode = (%d, %v), want (%d, true)", code, ok, messages.ErrSkew)
	}
	if _, ok := krbErrorCode([]byte{0x6b, 0x01}); ok {
		t.Error("expected ok=false for non-KRB-ERROR bytes")
	}
}

// TestResolveKDCAddrsIPLiteral verifies IP literals (IPv4 and IPv6) pass through
// without a DNS lookup, so no resolver is required.
func TestResolveKDCAddrsIPLiteral(t *testing.T) {
	for _, ip := range []string{"10.7.0.10", "::1", "fe80::1"} {
		addrs, err := resolveKDCAddrs(nil, ip)
		if err != nil {
			t.Fatalf("resolveKDCAddrs(%q): %v", ip, err)
		}
		if len(addrs) != 1 || addrs[0] != ip {
			t.Errorf("resolveKDCAddrs(%q) = %v, want [%q]", ip, addrs, ip)
		}
	}
}
