package authenticate

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

// TestUnmarshalMessageWithoutVersionOrMIC guards the defect: Unmarshal demanded
// 88 bytes -- the length of the fixed part with both Version and MIC -- although
// both are optional ([MS-NLMP] 2.2.1.3). A message carrying neither is 64 bytes
// before its payload and must parse.
func TestUnmarshalMessageWithoutVersionOrMIC(t *testing.T) {
	msg := AuthenticateMessage{
		NegotiateFlags:      flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_ANONYMOUS,
		LmChallengeResponse: []byte{0x00},
	}
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 65 {
		t.Fatalf("marshalled length = %d, want 65 (64-byte fixed part + 1-byte LM response)", len(raw))
	}

	var got AuthenticateMessage
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal of a 65-byte AUTHENTICATE: %v", err)
	}
	if got.Version != nil || got.NeedsMIC {
		t.Errorf("Version = %v, NeedsMIC = %v; want nil and false", got.Version, got.NeedsMIC)
	}
	if !bytes.Equal(got.LmChallengeResponse, []byte{0x00}) {
		t.Errorf("LmChallengeResponse = % x, want 00", got.LmChallengeResponse)
	}

	// The fixed part on its own, with no payload at all, is the shortest valid
	// message.
	empty := AuthenticateMessage{NegotiateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE}
	raw, err = empty.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 64 {
		t.Fatalf("marshalled length = %d, want 64", len(raw))
	}
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal of a 64-byte AUTHENTICATE: %v", err)
	}
}

// TestUnmarshalMessageWithVersionWithoutMIC covers the 72-byte fixed part: Version
// present, no MIC.
func TestUnmarshalMessageWithVersionWithoutMIC(t *testing.T) {
	v := version.DefaultVersion()
	msg := AuthenticateMessage{
		NegotiateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_VERSION,
		Version:        &v,
	}
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 72 {
		t.Fatalf("marshalled length = %d, want 72", len(raw))
	}

	var got AuthenticateMessage
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal of a 72-byte AUTHENTICATE: %v", err)
	}
	if got.Version == nil || *got.Version != v {
		t.Errorf("Version = %v, want %+v", got.Version, v)
	}
	if got.NeedsMIC {
		t.Error("NeedsMIC set on a message without a MIC")
	}
}

// TestAnonymousAuthenticateRoundTripsWithEmptyNames pins the case that exposed
// the defect: CreateAnonymousAuthenticateMessage with no domain and no workstation
// marshals to 73 bytes, which this package's own Unmarshal then rejected, so an
// acceptor refused the null session it was sent.
func TestAnonymousAuthenticateRoundTripsWithEmptyNames(t *testing.T) {
	for _, keyExch := range []bool{false, true} {
		negFlags := flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY
		if keyExch {
			negFlags |= flags.NTLMSSP_NEGOTIATE_KEY_EXCH
		}

		msg, err := CreateAnonymousAuthenticateMessage(newTestChallenge(negFlags), "", "")
		if err != nil {
			t.Fatalf("keyExch=%v: CreateAnonymousAuthenticateMessage: %v", keyExch, err)
		}
		raw, err := msg.Marshal()
		if err != nil {
			t.Fatalf("keyExch=%v: Marshal: %v", keyExch, err)
		}

		var got AuthenticateMessage
		if _, err := got.Unmarshal(raw); err != nil {
			t.Fatalf("keyExch=%v: Unmarshal of the %d-byte anonymous AUTHENTICATE: %v", keyExch, len(raw), err)
		}
		if !got.IsAnonymous() {
			t.Errorf("keyExch=%v: round trip lost NTLMSSP_NEGOTIATE_ANONYMOUS", keyExch)
		}
		if got.NeedsMIC {
			t.Errorf("keyExch=%v: NeedsMIC set on an anonymous AUTHENTICATE", keyExch)
		}
		if len(got.NtChallengeResponse) != 0 || !bytes.Equal(got.LmChallengeResponse, []byte{0x00}) {
			t.Errorf("keyExch=%v: NT=% x LM=% x, want empty and Z(1)", keyExch, got.NtChallengeResponse, got.LmChallengeResponse)
		}
		if !bytes.Equal(got.EncryptedRandomSessionKey, msg.EncryptedRandomSessionKey) {
			t.Errorf("keyExch=%v: EncryptedRandomSessionKey = % x, want % x", keyExch, got.EncryptedRandomSessionKey, msg.EncryptedRandomSessionKey)
		}
	}
}

// TestUnmarshalTruncatedMessagesDoNotPanic feeds every prefix of several valid
// messages -- with and without Version, with and without a MIC -- to Unmarshal.
// With the minimum lowered below 88, every read past the 64-byte fixed part rests
// on its own bounds check; none of the prefixes may panic, and every prefix
// shorter than the payload end must be rejected.
func TestUnmarshalTruncatedMessagesDoNotPanic(t *testing.T) {
	v := version.DefaultVersion()

	anonymous, err := CreateAnonymousAuthenticateMessage(newTestChallenge(flags.NTLMSSP_NEGOTIATE_UNICODE|flags.NTLMSSP_NEGOTIATE_KEY_EXCH), "", "WS")
	if err != nil {
		t.Fatalf("CreateAnonymousAuthenticateMessage: %v", err)
	}
	withMICNoVersion := &AuthenticateMessage{
		NegotiateFlags:      flags.NTLMSSP_NEGOTIATE_UNICODE,
		NeedsMIC:            true,
		UserName:            []byte("u\x00"),
		NtChallengeResponse: bytes.Repeat([]byte{0xaa}, 24),
	}
	withMICAndVersion := &AuthenticateMessage{
		NegotiateFlags:      flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_VERSION,
		Version:             &v,
		NeedsMIC:            true,
		DomainName:          []byte("d\x00"),
		NtChallengeResponse: bytes.Repeat([]byte{0xbb}, 24),
	}
	versionOnly := &AuthenticateMessage{
		NegotiateFlags:      flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_VERSION,
		Version:             &v,
		LmChallengeResponse: []byte{0x00},
	}

	for name, msg := range map[string]*AuthenticateMessage{
		"anonymous":           anonymous,
		"MIC without Version": withMICNoVersion,
		"MIC and Version":     withMICAndVersion,
		"Version only":        versionOnly,
	} {
		raw, err := msg.Marshal()
		if err != nil {
			t.Fatalf("%s: Marshal: %v", name, err)
		}
		var full AuthenticateMessage
		if _, err := full.Unmarshal(raw); err != nil {
			t.Fatalf("%s: Unmarshal of the full %d-byte message: %v", name, len(raw), err)
		}

		for n := 0; n < len(raw); n++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s: Unmarshal panicked on a %d-byte prefix: %v", name, n, r)
					}
				}()
				var got AuthenticateMessage
				if _, err := got.Unmarshal(raw[:n]); err == nil {
					t.Errorf("%s: Unmarshal accepted a %d-byte prefix of a %d-byte message", name, n, len(raw))
				}
			}()
		}
	}
}
