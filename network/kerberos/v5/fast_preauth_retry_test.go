package kerberos

import (
	"encoding/asn1"
	"fmt"
	"strings"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// A FAST-armored AS-REQ carries the client's credential proof as a
// PA-ENCRYPTED-CHALLENGE keyed by the client long-term key. The first attempt uses
// the default salt (UPPERCASE-REALM + account). An account whose salt is not that —
// a renamed domain or a pre-created account — makes the KDC answer with a
// FAST-armored KDC_ERR_PREAUTH_FAILED (24) that carries a PA-ETYPE-INFO2 naming the
// real salt. The client must apply that salt and retry, exactly as the non-FAST
// path does; before the fix it only retried on PREAUTH_REQUIRED (25) and so failed
// permanently against such accounts (observed live against a Windows Server 2025 DC
// where Administrator's salt is a legacy renamed-host salt).
//
// These tests drive the real getTGTFAST loop through an in-process mock KDC (the
// sendFunc seam), so they run in CI without a live DC.

const fastRetryPassword = "Passw0rd!"

// fastMockKDC emulates the KDC side of a FAST AS exchange enough to exercise the
// client's error handling. It reconstructs the armor key from each request (so it
// works with the client's random per-exchange subkey) and, per call, returns a
// FAST-armored KRB-ERROR with the configured error code and (optionally) a
// PA-ETYPE-INFO2 salt correction.
type fastMockKDC struct {
	t               *testing.T
	armorSessionKey []byte
	armorEType      int
	correctSalt     string // the salt the account really uses (advertised in ETYPE-INFO2)
	correctEType    int

	calls          int
	secondCallSalt string // the salt the client actually used on the retry, recovered from the challenge
}

// armorKeyFromRequest parses a FAST AS-REQ, decrypts the armor AP-REQ authenticator
// with the known armor session key to recover the client subkey, and derives the
// armor key the same way the client does.
func (k *fastMockKDC) armorKeyFromRequest(reqBytes []byte) (armorKey []byte, armorEType int, inner messages.KrbFastReq) {
	k.t.Helper()
	var asReq messages.ASReq
	if _, err := asReq.Unmarshal(reqBytes); err != nil {
		k.t.Fatalf("mock KDC: parse AS-REQ: %v", err)
	}
	var fastValue []byte
	for _, pa := range asReq.PAData {
		if pa.PADataType == messages.PAFXFast {
			fastValue = pa.PADataValue
		}
	}
	if fastValue == nil {
		k.t.Fatal("mock KDC: AS-REQ has no PA-FX-FAST")
	}
	// PA-FX-FAST request: CHOICE { armored-data [0] KrbFastArmoredReq }.
	var choice asn1.RawValue
	if _, err := asn1.Unmarshal(fastValue, &choice); err != nil {
		k.t.Fatalf("mock KDC: unwrap PA-FX-FAST: %v", err)
	}
	var armored messages.KrbFastArmoredReq
	if _, err := armored.Unmarshal(choice.Bytes); err != nil {
		k.t.Fatalf("mock KDC: parse KrbFastArmoredReq: %v", err)
	}

	var apReq messages.APReq
	if _, err := apReq.Unmarshal(armored.Armor.ArmorValue); err != nil {
		k.t.Fatalf("mock KDC: parse armor AP-REQ: %v", err)
	}
	authBytes, err := kerbcrypto.Decrypt(k.armorEType, k.armorSessionKey, kerbcrypto.KeyUsageAPReqAuthen, apReq.Authenticator.Cipher)
	if err != nil {
		k.t.Fatalf("mock KDC: decrypt armor authenticator: %v", err)
	}
	var auth messages.Authenticator
	if _, err := auth.Unmarshal(authBytes); err != nil {
		k.t.Fatalf("mock KDC: parse armor authenticator: %v", err)
	}
	if auth.SubKey == nil {
		k.t.Fatal("mock KDC: armor authenticator carries no subkey")
	}
	armorKey, armorEType, err = kerbcrypto.KRBFXCF2(
		auth.SubKey.KeyValue, auth.SubKey.KeyType,
		k.armorSessionKey, k.armorEType,
		fastPepperSubkeyArmor, fastPepperTicketArmor)
	if err != nil {
		k.t.Fatalf("mock KDC: derive armor key: %v", err)
	}

	fastReqBytes, err := kerbcrypto.Decrypt(armorEType, armorKey, kerbcrypto.KeyUsageFASTEnc, armored.EncFastReq.Cipher)
	if err != nil {
		k.t.Fatalf("mock KDC: decrypt enc-fast-req: %v", err)
	}
	if _, err := inner.Unmarshal(fastReqBytes); err != nil {
		k.t.Fatalf("mock KDC: parse KrbFastReq: %v", err)
	}
	return armorKey, armorEType, inner
}

// recoverChallengeSalt tries the advertised correctSalt against the inner
// PA-ENCRYPTED-CHALLENGE: if it decrypts, the client applied that salt on this
// attempt.
func (k *fastMockKDC) recoverChallengeSalt(armorKey []byte, armorEType int, inner messages.KrbFastReq) string {
	k.t.Helper()
	var challValue []byte
	for _, pa := range inner.PAData {
		if pa.PADataType == messages.PAEncryptedChallenge {
			challValue = pa.PADataValue
		}
	}
	if challValue == nil {
		k.t.Fatal("mock KDC: inner request has no PA-ENCRYPTED-CHALLENGE")
	}
	var enc messages.EncryptedData
	if _, err := asn1.Unmarshal(challValue, &enc); err != nil {
		k.t.Fatalf("mock KDC: parse encrypted challenge: %v", err)
	}
	clientKey, err := kerbcrypto.StringToKey(k.correctEType, fastRetryPassword, k.correctSalt, nil)
	if err != nil {
		k.t.Fatalf("mock KDC: client key: %v", err)
	}
	challengeKey, challEType, err := kerbcrypto.KRBFXCF2(
		armorKey, armorEType, clientKey, k.correctEType,
		fastPepperClientChallenge, fastPepperChallengeLongTem)
	if err != nil {
		k.t.Fatalf("mock KDC: derive challenge key: %v", err)
	}
	if _, err := kerbcrypto.Decrypt(challEType, challengeKey, kerbcrypto.KeyUsageEncChallengeClient, enc.Cipher); err != nil {
		return "" // the client did not use correctSalt on this attempt
	}
	return k.correctSalt
}

// armoredError builds a FAST-armored KRB-ERROR with the given code, optionally
// carrying a PA-ETYPE-INFO2 salt correction, encrypted under armorKey.
func (k *fastMockKDC) armoredError(armorKey []byte, armorEType, code int, withSaltInfo bool) []byte {
	k.t.Helper()
	innerErr := messages.KRBError{
		PVNO:      messages.KerberosV5,
		MsgType:   messages.MsgTypeError,
		ErrorCode: code,
		Realm:     testRealm,
		SName:     messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", testRealm}},
	}
	innerErrBytes, err := innerErr.Marshal()
	if err != nil {
		k.t.Fatalf("mock KDC: marshal inner error: %v", err)
	}
	paData := []messages.PAData{{PADataType: messages.PAFXError, PADataValue: innerErrBytes}}
	if withSaltInfo {
		info := messages.ETypeInfo2{{EType: k.correctEType, Salt: k.correctSalt}}
		infoBytes, err := info.Marshal()
		if err != nil {
			k.t.Fatalf("mock KDC: marshal ETYPE-INFO2: %v", err)
		}
		paData = append(paData, messages.PAData{PADataType: messages.PAETypeInfo2, PADataValue: infoBytes})
	}

	resp := messages.KrbFastResponse{PAData: paData}
	respBytes, err := resp.Marshal()
	if err != nil {
		k.t.Fatalf("mock KDC: marshal KrbFastResponse: %v", err)
	}
	encResp, err := kerbcrypto.Encrypt(armorEType, armorKey, kerbcrypto.KeyUsageFASTRep, respBytes)
	if err != nil {
		k.t.Fatalf("mock KDC: encrypt enc-fast-rep: %v", err)
	}
	armoredRep := messages.KrbFastArmoredRep{EncFastRep: messages.EncryptedData{EType: armorEType, Cipher: encResp}}
	armoredRepBytes, err := armoredRep.Marshal()
	if err != nil {
		k.t.Fatalf("mock KDC: marshal KrbFastArmoredRep: %v", err)
	}
	replyChoice, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: armoredRepBytes})
	if err != nil {
		k.t.Fatalf("mock KDC: wrap PA-FX-FAST reply: %v", err)
	}
	methodData := []messages.PAData{{PADataType: messages.PAFXFast, PADataValue: replyChoice}}
	edata, err := asn1.Marshal(methodData)
	if err != nil {
		k.t.Fatalf("mock KDC: marshal method-data: %v", err)
	}
	outer := messages.KRBError{
		PVNO:      messages.KerberosV5,
		MsgType:   messages.MsgTypeError,
		ErrorCode: code,
		Realm:     testRealm,
		SName:     messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", testRealm}},
		EData:     edata,
	}
	outerBytes, err := outer.Marshal()
	if err != nil {
		k.t.Fatalf("mock KDC: marshal outer error: %v", err)
	}
	return outerBytes
}

// newFASTRetryClient returns a client with a password credential and a FAST armor
// configured, pointed at mock through the sendFunc seam.
func newFASTRetryClient(t *testing.T, mock *fastMockKDC) *KerberosClient {
	t.Helper()
	c := fastArmoredClient(t, mock.armorSessionKey, mock.armorEType)
	return c
}

// TestFASTGetTGT_RetriesPreauthFailedWithSaltCorrection is the regression test for
// the fix: a FAST-armored PREAUTH_FAILED carrying a salt correction must trigger a
// retry whose challenge uses the corrected salt. Before the fix the client did not
// retry on error 24 and failed on the first response.
func TestFASTGetTGT_RetriesPreauthFailedWithSaltCorrection(t *testing.T) {
	mock := &fastMockKDC{
		t:               t,
		armorSessionKey: make([]byte, 32),
		armorEType:      messages.ETypeAES256CTSHMACSHA196,
		correctSalt:     "LEGACYHOST$alice", // not realm+username
		correctEType:    messages.ETypeAES256CTSHMACSHA196,
	}
	for i := range mock.armorSessionKey {
		mock.armorSessionKey[i] = byte(i + 1)
	}

	const sentinelCode = 60 // KRB_ERR_GENERIC: a distinct terminal code for the retry

	c := newFASTRetryClient(t, mock)
	c.sendFunc = func(realm string, msg []byte) ([]byte, error) {
		mock.calls++
		armorKey, armorEType, inner := mock.armorKeyFromRequest(msg)
		switch mock.calls {
		case 1:
			// First attempt used the default salt: reject with a salt correction.
			return mock.armoredError(armorKey, armorEType, int(messages.ErrPreauthFailed), true), nil
		default:
			// The retry: record whether the corrected salt was applied, then end
			// the exchange with a distinct terminal error.
			mock.secondCallSalt = mock.recoverChallengeSalt(armorKey, armorEType, inner)
			return mock.armoredError(armorKey, armorEType, sentinelCode, false), nil
		}
	}

	err := c.GetTGT()
	if err == nil {
		t.Fatal("expected the exchange to terminate with the sentinel error, got success")
	}
	if mock.calls != 2 {
		t.Fatalf("KDC was called %d times, want 2 (an optimistic attempt plus one corrected retry)", mock.calls)
	}
	if mock.secondCallSalt != mock.correctSalt {
		t.Errorf("retry challenge salt = %q, want the corrected %q", mock.secondCallSalt, mock.correctSalt)
	}
	if want := fmt.Sprintf("error %d", sentinelCode); !strings.Contains(err.Error(), want) {
		t.Errorf("final error = %q, want it to report the retry's response (%q)", err, want)
	}
}

// TestFASTGetTGT_NoRetryWithoutCorrection guards the retry bound: a PREAUTH_FAILED
// with no salt correction (a genuinely wrong credential) must not be retried, so a
// bad password cannot drive repeated failed logons against the KDC.
func TestFASTGetTGT_NoRetryWithoutCorrection(t *testing.T) {
	mock := &fastMockKDC{
		t:               t,
		armorSessionKey: make([]byte, 32),
		armorEType:      messages.ETypeAES256CTSHMACSHA196,
		correctSalt:     "LEGACYHOST$alice",
		correctEType:    messages.ETypeAES256CTSHMACSHA196,
	}
	for i := range mock.armorSessionKey {
		mock.armorSessionKey[i] = byte(i + 1)
	}

	c := newFASTRetryClient(t, mock)
	c.sendFunc = func(realm string, msg []byte) ([]byte, error) {
		mock.calls++
		armorKey, armorEType, _ := mock.armorKeyFromRequest(msg)
		return mock.armoredError(armorKey, armorEType, int(messages.ErrPreauthFailed), false), nil
	}

	if err := c.GetTGT(); err == nil {
		t.Fatal("expected PREAUTH_FAILED without a correction to be terminal")
	}
	if mock.calls != 1 {
		t.Fatalf("KDC was called %d times, want 1 (no retry without a salt correction)", mock.calls)
	}
}
