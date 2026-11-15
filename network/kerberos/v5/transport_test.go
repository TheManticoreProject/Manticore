package kerberos

import (
	"bytes"
	"context"
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

// startTCPKDC runs a loopback TCP "KDC" that reads one RFC 4120 length-framed
// request and replies with reply. It returns the listening port and a cleanup.
func startTCPKDC(t *testing.T, reply []byte) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				if _, err := readTCPFramed(c); err == nil {
					_ = writeTCPFramed(c, reply)
				}
			}(conn)
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, func() { ln.Close() }
}

// TestKDCSendAddrReturnsTCPProtocolError verifies that a complete KRB-ERROR
// received over TCP is returned to the caller. The KDC transport is TCP by
// default (Windows Vista+ / Impacket), so the loopback server speaks TCP.
func TestKDCSendAddrReturnsTCPProtocolError(t *testing.T) {
	want := mustMarshalKRBError(t, messages.ErrPreauthRequired)
	port, cleanup := startTCPKDC(t, want)
	defer cleanup()

	got, err := kdcSendAddr("127.0.0.1", port, []byte{0x01})
	if err != nil {
		t.Fatalf("kdcSendAddr: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("kdcSendAddr returned %x, want TCP KRB-ERROR %x", got, want)
	}
}

// TestKDCSendAddrUsesTCPNotUDP guards the IoC/malfunction fix: the client must
// use TCP and send no UDP datagram, so Kerberos works where UDP/88 is filtered.
// A UDP socket bound to the same port records any datagram; it must stay silent.
func TestKDCSendAddrUsesTCPNotUDP(t *testing.T) {
	want := mustMarshalKRBError(t, messages.ErrPreauthRequired)
	port, cleanup := startTCPKDC(t, want)
	defer cleanup()

	// Bind UDP on the same port number (independent namespace) to detect any
	// stray datagram the client might send.
	udp, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer udp.Close()
	gotUDP := make(chan struct{}, 1)
	go func() {
		buf := make([]byte, 64)
		if _, _, err := udp.ReadFromUDP(buf); err == nil {
			gotUDP <- struct{}{}
		}
	}()

	got, err := kdcSendAddr("127.0.0.1", port, []byte{0x01})
	if err != nil {
		t.Fatalf("kdcSendAddr over TCP: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("kdcSendAddr returned %x, want %x", got, want)
	}

	_ = udp.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	select {
	case <-gotUDP:
		t.Fatal("client sent a UDP datagram; the KDC transport must be TCP by default")
	case <-time.After(250 * time.Millisecond):
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

func TestKDCSendAddrHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := kdcSendAddrContext(ctx, "127.0.0.1", 88, []byte{0x01}, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("kdcSendAddrContext error = %v, want context.Canceled", err)
	}
}

func TestKDCSendAddrHonorsTimeout(t *testing.T) {
	// A TCP listener that accepts the connection but never replies, so the read
	// blocks until the operation deadline fires.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
		}
	}()

	start := time.Now()
	_, err = kdcSendAddrContext(context.Background(), "127.0.0.1", ln.Addr().(*net.TCPAddr).Port, []byte{0x01}, 25*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("kdcSendAddrContext error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("configured timeout took %s", elapsed)
	}
}

func TestKerberosClientOperationControls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := NewClient("alice", "corp.local", "127.0.0.1").WithContext(ctx).WithTimeout(125 * time.Millisecond)

	gotCtx, gotTimeout := c.contextAndTimeout()
	if gotCtx != ctx || gotTimeout != 125*time.Millisecond {
		t.Fatalf("contextAndTimeout = (%v, %s), want configured context and 125ms", gotCtx, gotTimeout)
	}
	cancel()
	if _, err := c.sendToRealm(c.realm, []byte{0x01}); !errors.Is(err, context.Canceled) {
		t.Fatalf("client transport error = %v, want context.Canceled", err)
	}
}
