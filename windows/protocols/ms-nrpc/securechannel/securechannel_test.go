package securechannel

// White-box tests for the Netlogon secure channel. The rolling test drives an independent
// server mirror of [MS-NRPC] 3.1.4.5; the Establish tests use a scripted invoker that answers
// the handshake opnums with NDR responses built from mirror structs (the interface package's
// response types are unexported, so the mock cannot fill them directly).

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
	nrpccrypto "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc/crypto"
)

var scSessionKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

// TestSecureChannelRolling drives NextAuthenticator/VerifyResponseAuthenticator across
// several calls against an independent server mirror of 3.1.4.5, confirming the client and
// server stored-credential accumulators stay in lockstep.
func TestSecureChannelRolling(t *testing.T) {
	seed := msnrpc.NETLOGON_CREDENTIAL{0x7e, 0x0a, 0xa0, 0x52, 0x63, 0xd9, 0xe4, 0xee}
	timestamps := []uint32{0x1000, 0x1005, 0x100a, 0x2000}
	idx := 0
	sc := &SecureChannel{
		sessionKey:       scSessionKey,
		aes:              true,
		clientStoredCred: seed,
		now:              func() time.Time { t := time.Unix(int64(timestamps[idx]), 0); idx++; return t },
	}

	serverStored := seed // server seeds its stored credential identically
	for _, ts := range timestamps {
		auth := sc.NextAuthenticator()
		if auth.Timestamp != ts {
			t.Fatalf("timestamp = %#x, want %#x", auth.Timestamp, ts)
		}
		// Server side ([MS-NRPC] 3.1.4.5 step 2): stored += ts, verify, stored += 1, respond.
		serverStored = nrpccrypto.AddToCredential(serverStored, ts)
		if want := nrpccrypto.ComputeNetlogonCredentialAES(serverStored, scSessionKey); want != auth.Credential {
			t.Fatalf("ts %#x: client credential %x, server expected %x", ts, auth.Credential, want)
		}
		serverStored = nrpccrypto.AddToCredential(serverStored, 1)
		resp := msnrpc.NETLOGON_AUTHENTICATOR{
			Credential: nrpccrypto.ComputeNetlogonCredentialAES(serverStored, scSessionKey),
			Timestamp:  ts,
		}
		if err := sc.VerifyResponseAuthenticator(resp); err != nil {
			t.Fatalf("ts %#x: VerifyResponseAuthenticator: %v", ts, err)
		}
	}
}

// TestVerifyResponseRejectsForgedAuthenticator confirms a bad server credential is rejected.
func TestVerifyResponseRejectsForgedAuthenticator(t *testing.T) {
	sc := &SecureChannel{
		sessionKey:       scSessionKey,
		aes:              true,
		clientStoredCred: msnrpc.NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8},
		now:              func() time.Time { return time.Unix(0x3000, 0) },
	}
	sc.NextAuthenticator()
	if err := sc.VerifyResponseAuthenticator(msnrpc.NETLOGON_AUTHENTICATOR{}); err == nil {
		t.Fatal("VerifyResponseAuthenticator accepted an all-zero server authenticator")
	}
}

// reqChallengeResp / authenticate3Resp mirror the interface package's (unexported) response
// layouts so the mock can produce a matching NDR response stub.
type reqChallengeResp struct {
	ServerChallenge msnrpc.NETLOGON_CREDENTIAL
	Status          ndr.DWORD `ndr:"retval"`
}

type authenticate3Resp struct {
	ServerCredential msnrpc.NETLOGON_CREDENTIAL
	NegotiateFlags   ndr.DWORD
	AccountRid       ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// scriptInvoker answers the two handshake opnums from canned data, records the opnums seen
// and the Authenticate3 request stub, and can forge the server credential or fail
// ReqChallenge. The server credential is computed with the same cipher suite (aes) the client
// will use, so it verifies unless forged.
type scriptInvoker struct {
	serverChallenge msnrpc.NETLOGON_CREDENTIAL
	sessionKey      [16]byte
	aes             bool
	echoFlags       uint32
	reqStatus       uint32
	forgeServer     bool
	opnums          []uint16
	auth3Stub       []byte
}

func (s *scriptInvoker) serverCred() msnrpc.NETLOGON_CREDENTIAL {
	if s.aes {
		return nrpccrypto.ComputeNetlogonCredentialAES(s.serverChallenge, s.sessionKey)
	}
	return nrpccrypto.ComputeNetlogonCredential(s.serverChallenge, s.sessionKey)
}

func (s *scriptInvoker) Invoke(in ndr.Call, out any) error {
	s.opnums = append(s.opnums, in.Opnum())
	stub, err := ndr.Request(in)
	if err != nil {
		return err
	}
	switch in.Opnum() {
	case netlogon.OpnumNetrServerReqChallenge:
		b, err := ndr.Marshal(&reqChallengeResp{ServerChallenge: s.serverChallenge, Status: ndr.DWORD(s.reqStatus)})
		if err != nil {
			return err
		}
		return ndr.Response(b, out)
	case netlogon.OpnumNetrServerAuthenticate3:
		s.auth3Stub = stub
		cred := s.serverCred()
		if s.forgeServer {
			cred[0] ^= 0xff
		}
		b, err := ndr.Marshal(&authenticate3Resp{ServerCredential: cred, NegotiateFlags: ndr.DWORD(s.echoFlags), Status: ndr.DWORD(netlogon.StatusSuccess)})
		if err != nil {
			return err
		}
		return ndr.Response(b, out)
	}
	return fmt.Errorf("scriptInvoker: unexpected opnum %d", in.Opnum())
}

// establishFixture builds a scripted invoker and config for a fixed client/server challenge
// and password, for the AES suite (aes=true) or the legacy strong-key suite (aes=false).
func establishFixture(aes, forge bool, reqStatus uint32) (*scriptInvoker, SecureChannelConfig, msnrpc.NETLOGON_CREDENTIAL, [16]byte) {
	clientChallenge := msnrpc.NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8}
	serverChallenge := msnrpc.NETLOGON_CREDENTIAL{8, 7, 6, 5, 4, 3, 2, 1}
	const password = "Machine$Pass1"
	var sk [16]byte
	flags := DefaultNegotiateFlags // has NegotiateAES
	if aes {
		sk = nrpccrypto.ComputeSessionKeyAES(password, nil, clientChallenge, serverChallenge)
	} else {
		sk = nrpccrypto.ComputeSessionKeyStrongKey(password, nil, clientChallenge, serverChallenge)
		flags = DefaultNegotiateFlags &^ netlogon.NegotiateAES // clear AES -> strong-key suite
	}
	inv := &scriptInvoker{serverChallenge: serverChallenge, sessionKey: sk, aes: aes, echoFlags: flags, reqStatus: reqStatus, forgeServer: forge}
	cfg := SecureChannelConfig{
		PrimaryName:       "DC01",
		ComputerName:      "WS01",
		AccountName:       "WS01$",
		SecureChannelType: msnrpc.WorkstationSecureChannel,
		Password:          password,
		NegotiateFlags:    flags,
		Rand:              bytes.NewReader(clientChallenge[:]),
	}
	return inv, cfg, clientChallenge, sk
}

// TestEstablishHappyPathAES verifies the handshake sequence, the derived session key, the
// seeded stored credential, and the credential the client sent in NetrServerAuthenticate3.
func TestEstablishHappyPathAES(t *testing.T) {
	inv, cfg, clientChallenge, sk := establishFixture(true, false, netlogon.StatusSuccess)

	sc, err := Establish(inv, cfg)
	if err != nil {
		t.Fatalf("Establish: %v", err)
	}
	if len(inv.opnums) != 2 || inv.opnums[0] != netlogon.OpnumNetrServerReqChallenge || inv.opnums[1] != netlogon.OpnumNetrServerAuthenticate3 {
		t.Fatalf("opnum sequence = %v, want [4 26]", inv.opnums)
	}
	if sc.SessionKey() != sk {
		t.Errorf("session key = %x, want %x", sc.SessionKey(), sk)
	}
	if !sc.UsesAES() {
		t.Error("UsesAES() = false, want true")
	}
	wantSeed := nrpccrypto.ComputeNetlogonCredentialAES(clientChallenge, sk)
	if sc.clientStoredCred != wantSeed {
		t.Errorf("seeded stored credential = %x, want %x", sc.clientStoredCred, wantSeed)
	}
	// The Authenticate3 request must carry the client credential and the account name.
	if !bytes.Contains(inv.auth3Stub, wantSeed[:]) {
		t.Errorf("client credential %x not found in Authenticate3 stub", wantSeed)
	}
	if !bytes.Contains(inv.auth3Stub, []byte{'W', 0, 'S', 0, '0', 0, '1', 0, '$', 0}) {
		t.Error("account name WS01$ not found in Authenticate3 stub")
	}
}

// TestEstablishHappyPathRC4 exercises the legacy strong-key suite selection end to end
// (offline): the session key is strong-key-derived and the credential is DES-based.
func TestEstablishHappyPathRC4(t *testing.T) {
	inv, cfg, _, sk := establishFixture(false, false, netlogon.StatusSuccess)
	sc, err := Establish(inv, cfg)
	if err != nil {
		t.Fatalf("Establish (RC4 suite): %v", err)
	}
	if sc.UsesAES() {
		t.Error("UsesAES() = true, want false for the strong-key suite")
	}
	if sc.SessionKey() != sk {
		t.Errorf("session key = %x, want %x", sc.SessionKey(), sk)
	}
}

// TestEstablishRejectsForgedServerCredential confirms Establish fails when the server does
// not prove knowledge of the shared secret.
func TestEstablishRejectsForgedServerCredential(t *testing.T) {
	inv, cfg, _, _ := establishFixture(true, true, netlogon.StatusSuccess)
	if _, err := Establish(inv, cfg); err == nil {
		t.Fatal("Establish accepted a forged server credential")
	}
}

// TestEstablishReqChallengeStatusError confirms a non-success ReqChallenge status aborts the
// handshake before NetrServerAuthenticate3 is called.
func TestEstablishReqChallengeStatusError(t *testing.T) {
	inv, cfg, _, _ := establishFixture(true, false, netlogon.StatusAccessDenied)
	if _, err := Establish(inv, cfg); err == nil {
		t.Fatal("Establish ignored a failed NetrServerReqChallenge status")
	}
	if len(inv.opnums) != 1 || inv.opnums[0] != netlogon.OpnumNetrServerReqChallenge {
		t.Fatalf("opnums = %v, want just [4] (no Authenticate3 after failure)", inv.opnums)
	}
}
