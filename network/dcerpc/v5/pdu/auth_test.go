package pdu

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
)

// TestSecTrailerRoundTrip verifies the 8-byte sec_trailer marshals and unmarshals
// to identical values.
func TestSecTrailerRoundTrip(t *testing.T) {
	st := SecTrailer{
		AuthType:      AuthTypeNTLMSSP,
		AuthLevel:     AuthLevelPktPrivacy,
		AuthPadLength: 3,
		AuthReserved:  0,
		AuthContextID: 0x12345678,
	}
	raw := st.Marshal()
	if len(raw) != SecTrailerSize {
		t.Fatalf("marshaled sec_trailer is %d bytes, want %d", len(raw), SecTrailerSize)
	}
	var got SecTrailer
	if err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != st {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, st)
	}
}

// TestAuth3RoundTrip verifies an auth3 PDU marshals and parses back, with auth_length
// counting only the auth_value and the sec_trailer carried intact.
func TestAuth3RoundTrip(t *testing.T) {
	authValue := []byte("NTLM-AUTHENTICATE-token-bytes")
	a := &Auth3{
		SecTrailer: SecTrailer{AuthType: AuthTypeNTLMSSP, AuthLevel: AuthLevelPktPrivacy, AuthContextID: 1},
		AuthValue:  authValue,
	}
	a.Header = NewHeader(PacketTypeAuth3, PFCFirstFrag|PFCLastFrag, 1)

	raw, err := a.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	hdr, err := PeekHeader(raw)
	if err != nil {
		t.Fatalf("PeekHeader: %v", err)
	}
	if hdr.PacketType != PacketTypeAuth3 {
		t.Errorf("packet type = %s, want auth3", hdr.PacketType)
	}
	if int(hdr.AuthLength) != len(authValue) {
		t.Errorf("auth_length = %d, want %d", hdr.AuthLength, len(authValue))
	}
	if int(hdr.FragLength) != len(raw) {
		t.Errorf("frag_length = %d, want %d", hdr.FragLength, len(raw))
	}

	var got Auth3
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !bytes.Equal(got.AuthValue, authValue) {
		t.Errorf("auth_value = %q, want %q", got.AuthValue, authValue)
	}
	if got.SecTrailer != a.SecTrailer {
		t.Errorf("sec_trailer = %+v, want %+v", got.SecTrailer, a.SecTrailer)
	}
}

// TestBindAuthVerifier verifies a bind carrying an auth_value sets auth_length, aligns
// the sec_trailer to a 4-byte boundary, and that ExtractAuthVerifier recovers the
// sec_trailer and token from the marshaled PDU.
func TestBindAuthVerifier(t *testing.T) {
	authValue := []byte("NTLM-NEGOTIATE")
	bind := &Bind{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []ContextElement{{
			ContextID:        0,
			AbstractSyntax:   syntax.NDRTransferSyntax(),
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
		SecTrailer: SecTrailer{AuthType: AuthTypeNTLMSSP, AuthLevel: AuthLevelPktPrivacy, AuthContextID: 0},
		AuthValue:  authValue,
	}
	bind.Header = NewHeader(PacketTypeBind, PFCFirstFrag|PFCLastFrag, 1)

	raw, err := bind.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	hdr, err := PeekHeader(raw)
	if err != nil {
		t.Fatalf("PeekHeader: %v", err)
	}
	if int(hdr.AuthLength) != len(authValue) {
		t.Errorf("auth_length = %d, want %d", hdr.AuthLength, len(authValue))
	}

	st, gotValue, err := ExtractAuthVerifier(raw)
	if err != nil {
		t.Fatalf("ExtractAuthVerifier: %v", err)
	}
	if st == nil {
		t.Fatal("ExtractAuthVerifier returned no sec_trailer")
	}
	if st.AuthType != AuthTypeNTLMSSP || st.AuthLevel != AuthLevelPktPrivacy {
		t.Errorf("sec_trailer = %+v", st)
	}
	if !bytes.Equal(gotValue, authValue) {
		t.Errorf("auth_value = %q, want %q", gotValue, authValue)
	}

	// The sec_trailer must begin on a 4-byte boundary from the start of the PDU.
	trailerStart := len(raw) - len(authValue) - SecTrailerSize
	if trailerStart%4 != 0 {
		t.Errorf("sec_trailer starts at offset %d, not 4-byte aligned", trailerStart)
	}
	// The recorded auth_pad_length is the padding between the body and the trailer,
	// which alignment keeps below 4.
	if st.AuthPadLength >= 4 {
		t.Errorf("auth_pad_length = %d, want < 4", st.AuthPadLength)
	}
}

// TestExtractAuthVerifierNone confirms a PDU without an auth verifier yields nil.
func TestExtractAuthVerifierNone(t *testing.T) {
	bind := &Bind{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []ContextElement{{
			ContextID:        0,
			AbstractSyntax:   syntax.NDRTransferSyntax(),
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
	}
	bind.Header = NewHeader(PacketTypeBind, PFCFirstFrag|PFCLastFrag, 1)
	raw, err := bind.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	st, val, err := ExtractAuthVerifier(raw)
	if err != nil {
		t.Fatalf("ExtractAuthVerifier: %v", err)
	}
	if st != nil || val != nil {
		t.Errorf("expected no auth verifier, got sec_trailer=%v value=%v", st, val)
	}
}
