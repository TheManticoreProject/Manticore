package spnego

import (
	"bytes"
	"encoding/asn1"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/security"
)

// TestMarshalMechTypeListOmitsTheContextTag pins the input a mechListMIC is taken
// over. RFC 4178 section 5 is explicit that the message is the DER encoding of
// MechTypeList and NOT of "[0] MechTypeList", so a leading 0xA0 context tag here
// would make every MIC disagree with the peer's.
func TestMarshalMechTypeListOmitsTheContextTag(t *testing.T) {
	der, err := MarshalMechTypeList([]asn1.ObjectIdentifier{NtlmOID})
	if err != nil {
		t.Fatalf("MarshalMechTypeList: %v", err)
	}
	if len(der) == 0 {
		t.Fatal("MarshalMechTypeList returned no bytes")
	}
	if der[0] == 0xA0 {
		t.Errorf("MechTypeList DER starts with the [0] context tag: % x", der)
	}
	if der[0] != 0x30 {
		t.Errorf("MechTypeList DER starts with %#02x, want 0x30 (SEQUENCE)", der[0])
	}

	// The NTLMSSP OID must appear in the encoding.
	want, err := asn1.Marshal(NtlmOID)
	if err != nil {
		t.Fatalf("marshal OID: %v", err)
	}
	if !bytes.Contains(der, want) {
		t.Errorf("MechTypeList DER % x does not contain the NTLMSSP OID % x", der, want)
	}
}

// authContextForMIC builds a context in the state it reaches after a successful
// AUTHENTICATE: a session key derived and the advertised mech list retained.
func authContextForMIC(t *testing.T) *AuthContext {
	t.Helper()

	der, err := MarshalMechTypeList([]asn1.ObjectIdentifier{NtlmOID})
	if err != nil {
		t.Fatalf("MarshalMechTypeList: %v", err)
	}
	return &AuthContext{
		SessionKey:  []byte("0123456789abcdef"),
		MechListDER: der,
		authenticateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE |
			flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			flags.NTLMSSP_NEGOTIATE_SIGN |
			flags.NTLMSSP_NEGOTIATE_KEY_EXCH,
	}
}

// TestComputeMechListMICIsAnNTLMSignature checks the MIC is a well-formed
// NTLMSSP_MESSAGE_SIGNATURE over the advertised mech list, at sequence number 0 —
// it is the first message protected on the new session.
func TestComputeMechListMICIsAnNTLMSignature(t *testing.T) {
	ctx := authContextForMIC(t)

	mic, err := ctx.computeMechListMIC(ctx.authenticateFlags)
	if err != nil {
		t.Fatalf("computeMechListMIC: %v", err)
	}
	if len(mic) != security.SignatureSize {
		t.Fatalf("mechListMIC is %d bytes, want %d", len(mic), security.SignatureSize)
	}

	// Version 1, little-endian, then an 8-byte checksum, then the sequence number.
	if got := uint32(mic[0]) | uint32(mic[1])<<8 | uint32(mic[2])<<16 | uint32(mic[3])<<24; got != 1 {
		t.Errorf("signature version = %d, want 1", got)
	}
	if got := uint32(mic[12]) | uint32(mic[13])<<8 | uint32(mic[14])<<16 | uint32(mic[15])<<24; got != 0 {
		t.Errorf("sequence number = %d, want 0", got)
	}

	// Computing it again from an equivalent context must give the same bytes: the
	// MIC is a function of the key and the mech list, not of call order.
	again, err := authContextForMIC(t).computeMechListMIC(ctx.authenticateFlags)
	if err != nil {
		t.Fatalf("computeMechListMIC (second context): %v", err)
	}
	if !bytes.Equal(mic, again) {
		t.Errorf("mechListMIC is not deterministic:\n first % x\nsecond % x", mic, again)
	}
}

// TestComputeMechListMICCoversTheMechList checks the MIC actually depends on the
// mech list — a MIC that ignored its input would be worthless as negotiation
// protection.
func TestComputeMechListMICCoversTheMechList(t *testing.T) {
	ctx := authContextForMIC(t)
	base, err := ctx.computeMechListMIC(ctx.authenticateFlags)
	if err != nil {
		t.Fatalf("computeMechListMIC: %v", err)
	}

	altered := authContextForMIC(t)
	der, err := MarshalMechTypeList([]asn1.ObjectIdentifier{NtlmOID, KerberosOID})
	if err != nil {
		t.Fatalf("MarshalMechTypeList: %v", err)
	}
	altered.MechListDER = der

	other, err := altered.computeMechListMIC(altered.authenticateFlags)
	if err != nil {
		t.Fatalf("computeMechListMIC (altered list): %v", err)
	}
	if bytes.Equal(base, other) {
		t.Error("the mechListMIC is unchanged by a different mech list, so it protects nothing")
	}
}

// TestComputeMechListMICWithoutASessionIsSkipped checks the exchange still works
// where no MIC can be produced: a mechanism that derives no key, or a context with
// no recorded mech list. Emitting none is valid whenever the peer does not require
// one, and is what the exchange did before.
func TestComputeMechListMICWithoutASessionIsSkipped(t *testing.T) {
	noKey := authContextForMIC(t)
	noKey.SessionKey = nil
	if mic, err := noKey.computeMechListMIC(noKey.authenticateFlags); err != nil || mic != nil {
		t.Errorf("with no session key: got (% x, %v), want (nil, nil)", mic, err)
	}

	noList := authContextForMIC(t)
	noList.MechListDER = nil
	if mic, err := noList.computeMechListMIC(noList.authenticateFlags); err != nil || mic != nil {
		t.Errorf("with no mech list: got (% x, %v), want (nil, nil)", mic, err)
	}
}

// TestVerifyServerMechListMICAcceptsAbsentAndRejectsWrong covers the receiving
// side: an omitted MIC is valid, and one that does not verify is an error.
func TestVerifyServerMechListMICAcceptsAbsentAndRejectsWrong(t *testing.T) {
	ctx := authContextForMIC(t)

	// No token at all, and a token with no mechListMIC, both verify trivially.
	if err := ctx.VerifyServerMechListMIC(nil); err != nil {
		t.Errorf("empty token: %v", err)
	}

	withoutMIC := NegTokenResp{NegState: NegStateAcceptCompleted}
	tokenNoMIC, err := withoutMIC.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := ctx.VerifyServerMechListMIC(tokenNoMIC); err != nil {
		t.Errorf("token without a mechListMIC: %v", err)
	}

	// A MIC of the right length but the wrong bytes must be rejected.
	bad := NegTokenResp{NegState: NegStateAcceptCompleted, MechListMIC: make([]byte, security.SignatureSize)}
	tokenBad, err := bad.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := ctx.VerifyServerMechListMIC(tokenBad); err == nil {
		t.Error("a zeroed mechListMIC was accepted, want a verification failure")
	}
}
