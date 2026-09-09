package authenticate_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
)

const (
	kxUser     = "Administrator"
	kxPassword = "pass"
	kxDomain   = "TMP"
)

// kxChallenge builds a CHALLENGE with a server TargetInfo carrying an
// MsvAvTimestamp, as a Windows domain controller sends.
func kxChallenge(t *testing.T) *challenge.ChallengeMessage {
	t.Helper()

	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, (uint64(time.Now().Unix())+116444736000)*10000000)

	info, err := targetinfo.BuildServerTargetInfo("DC01", kxDomain, "dc01.tmp.local", "tmp.local", ts)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}

	msg := &challenge.ChallengeMessage{
		NegotiateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE |
			flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			flags.NTLMSSP_NEGOTIATE_TARGET_INFO,
		TargetInfo: info,
	}
	copy(msg.ServerChallenge[:], []byte{1, 2, 3, 4, 5, 6, 7, 8})
	return msg
}

// sessionBaseKeyOf recomputes the SessionBaseKey the server would derive:
// HMAC-MD5(NTOWFv2, NTProofStr), where NTProofStr is the leading 16 bytes of the
// NTLMv2 response ([MS-NLMP] 2.2.2.8).
func sessionBaseKeyOf(t *testing.T, ntChallengeResponse []byte) []byte {
	t.Helper()
	if len(ntChallengeResponse) < 16 {
		t.Fatalf("NtChallengeResponse is %d bytes, too short to contain an NTProofStr", len(ntChallengeResponse))
	}
	mac := hmac.New(md5.New, ntlmv2.NTOWFv2(kxPassword, kxUser, kxDomain))
	mac.Write(ntChallengeResponse[:16])
	return mac.Sum(nil)
}

// TestKeyExchangeWrapsARandomSessionKey guards the defect: key exchange was stubbed
// out, so the exported session key was the SessionBaseKey and
// EncryptedRandomSessionKey was always empty.
//
// It verifies the exchange cryptographically rather than structurally: the wrapped
// key must decrypt, under the KeyExchangeKey the server independently derives, to
// exactly the key the client kept for signing ([MS-NLMP] 3.1.5.1.2).
func TestKeyExchangeWrapsARandomSessionKey(t *testing.T) {
	ch := kxChallenge(t)
	ch.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_KEY_EXCH

	msg, err := authenticate.CreateAuthenticateMessage(ch, kxUser, kxPassword, kxDomain, "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_KEY_EXCH == 0 {
		t.Error("AUTHENTICATE does not assert NTLMSSP_NEGOTIATE_KEY_EXCH")
	}
	if len(msg.EncryptedRandomSessionKey) != 16 {
		t.Fatalf("EncryptedRandomSessionKey is %d bytes, want 16", len(msg.EncryptedRandomSessionKey))
	}
	if len(msg.SessionKey) != 16 {
		t.Fatalf("SessionKey is %d bytes, want 16", len(msg.SessionKey))
	}

	// For NTLMv2 the KeyExchangeKey is the SessionBaseKey (MS-NLMP 3.4.5.1).
	keyExchangeKey := sessionBaseKeyOf(t, msg.NtChallengeResponse)

	if bytes.Equal(msg.SessionKey, keyExchangeKey) {
		t.Error("SessionKey equals the SessionBaseKey: no independent key was generated")
	}

	cipher, err := rc4.NewCipher(keyExchangeKey)
	if err != nil {
		t.Fatalf("rc4.NewCipher: %v", err)
	}
	unwrapped := make([]byte, 16)
	cipher.XORKeyStream(unwrapped, msg.EncryptedRandomSessionKey)

	if !bytes.Equal(unwrapped, msg.SessionKey) {
		t.Errorf("unwrapping EncryptedRandomSessionKey under the KeyExchangeKey gave % x, want the retained SessionKey % x",
			unwrapped, msg.SessionKey)
	}
}

// TestWithoutKeyExchangeSessionKeyIsTheBaseKey pins the other branch: a server that
// does not offer key exchange must not be sent a wrapped key, and the exported key
// is then the KeyExchangeKey itself.
func TestWithoutKeyExchangeSessionKeyIsTheBaseKey(t *testing.T) {
	ch := kxChallenge(t)
	ch.NegotiateFlags &^= flags.NTLMSSP_NEGOTIATE_KEY_EXCH

	msg, err := authenticate.CreateAuthenticateMessage(ch, kxUser, kxPassword, kxDomain, "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_KEY_EXCH != 0 {
		t.Error("AUTHENTICATE asserts NTLMSSP_NEGOTIATE_KEY_EXCH although the CHALLENGE did not offer it")
	}
	if len(msg.EncryptedRandomSessionKey) != 0 {
		t.Errorf("EncryptedRandomSessionKey is %d bytes, want empty when key exchange was not negotiated", len(msg.EncryptedRandomSessionKey))
	}
	if want := sessionBaseKeyOf(t, msg.NtChallengeResponse); !bytes.Equal(msg.SessionKey, want) {
		t.Errorf("SessionKey = % x, want the SessionBaseKey % x", msg.SessionKey, want)
	}
}
