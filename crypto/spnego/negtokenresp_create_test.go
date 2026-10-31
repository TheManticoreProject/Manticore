package spnego_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
)

// TestCreateNegTokenRespEncoding checks that CreateNegTokenResp emits a continuation
// token as RFC 4178 4.2 defines it: the negTokenResp [1] choice of
// NegotiationToken, not GSS-wrapped, carrying EXPLICIT-tagged fields, and that the
// result decodes with NegTokenResp.Unmarshal and ExtractNTLMToken.
func TestCreateNegTokenRespEncoding(t *testing.T) {
	ntlmToken := []byte("NTLMSSP\x00\x02\x00\x00\x00challenge")

	token, err := spnego.CreateNegTokenResp(spnego.NegStateAcceptIncomplete, spnego.NtlmOID, ntlmToken)
	if err != nil {
		t.Fatalf("CreateNegTokenResp: %v", err)
	}

	// Byte-identical to the token the acceptor builds from NegTokenResp.Marshal.
	inner, err := spnego.NewNegTokenResp(spnego.NegStateAcceptIncomplete, spnego.NtlmOID, ntlmToken).Marshal()
	if err != nil {
		t.Fatalf("NegTokenResp.Marshal: %v", err)
	}
	want, err := (&spnego.SecurityBlob{Data: inner}).Marshal()
	if err != nil {
		t.Fatalf("SecurityBlob.Marshal: %v", err)
	}
	if !bytes.Equal(token, want) {
		t.Fatalf("CreateNegTokenResp =\n  % x\nwant\n  % x", token, want)
	}

	var blob spnego.SecurityBlob
	if _, err := blob.Unmarshal(token); err != nil {
		t.Fatalf("token is not a [1] negTokenResp: %v", err)
	}
	var resp spnego.NegTokenResp
	if _, err := resp.Unmarshal(blob.Data); err != nil {
		t.Fatalf("NegTokenResp.Unmarshal: %v", err)
	}
	if resp.NegState != spnego.NegStateAcceptIncomplete {
		t.Errorf("negState = %v, want %v", resp.NegState, spnego.NegStateAcceptIncomplete)
	}
	if !resp.SupportedMech.Equal(spnego.NtlmOID) {
		t.Errorf("supportedMech = %v, want %v", resp.SupportedMech, spnego.NtlmOID)
	}
	if !bytes.Equal(resp.ResponseToken, ntlmToken) {
		t.Errorf("responseToken = % x, want % x", resp.ResponseToken, ntlmToken)
	}

	extracted, err := spnego.ExtractNTLMToken(token)
	if err != nil {
		t.Fatalf("ExtractNTLMToken: %v", err)
	}
	if !bytes.Equal(extracted, ntlmToken) {
		t.Errorf("ExtractNTLMToken = % x, want % x", extracted, ntlmToken)
	}
}
