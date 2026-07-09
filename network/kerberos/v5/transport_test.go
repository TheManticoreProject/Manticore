package kerberos

import (
	"errors"
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
// datagram, and any KRB-ERROR reply all force a TCP retry; a normal AS-REP does
// not.
func TestShouldRetryOverTCP(t *testing.T) {
	krbErr := mustMarshalKRBError(t, messages.ErrResponseTooBig)

	tests := []struct {
		name string
		resp []byte
		err  error
		want bool
	}{
		{"udp error", nil, errors.New("timeout"), true},
		{"empty datagram", []byte{}, nil, true},
		{"krb-error reply", krbErr, nil, true},
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
