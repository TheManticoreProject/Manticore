package functions

// White-box tests for the Netlogon secure channel. Being in package functions lets the
// scripted invoker populate the unexported response structs directly and lets the rolling
// test drive an independent server mirror of [MS-NRPC] 3.1.4.5. The authenticator vector is
// pinned against impacket's ComputeNetlogonAuthenticatorAES.

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

var scSessionKey = [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

// TestComputeNetlogonAuthenticatorAESVector pins the authenticator arithmetic to the
// impacket vector: stored 0x11*8, timestamp 0x11223344, session key 00..0f.
func TestComputeNetlogonAuthenticatorAESVector(t *testing.T) {
	stored := msnrpc.NETLOGON_CREDENTIAL{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}
	auth := ComputeNetlogonAuthenticatorAES(stored, 0x11223344, scSessionKey)
	if got := hex.EncodeToString(auth.Credential[:]); got != "934de307ed961bf9" {
		t.Errorf("credential = %s, want 934de307ed961bf9", got)
	}
	if auth.Timestamp != 0x11223344 {
		t.Errorf("timestamp = %#x, want 0x11223344", auth.Timestamp)
	}
}

// TestAddToCredential pins the low-32-bit little-endian add with overflow ignored and the
// high 4 bytes left untouched ([MS-NRPC] 3.1.4.5).
func TestAddToCredential(t *testing.T) {
	// 0xFFFFFFFF + 2 wraps to 0x00000001 in the low DWORD; high DWORD unchanged.
	in := msnrpc.NETLOGON_CREDENTIAL{0xff, 0xff, 0xff, 0xff, 0xaa, 0xbb, 0xcc, 0xdd}
	got := addToCredential(in, 2)
	want := msnrpc.NETLOGON_CREDENTIAL{0x01, 0x00, 0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd}
	if got != want {
		t.Errorf("addToCredential = %x, want %x", got, want)
	}
}

// TestSecureChannelRolling drives NextAuthenticator/VerifyResponseAuthenticator across
// several calls against an independent server mirror of 3.1.4.5, confirming the client and
// server stored-credential accumulators stay in lockstep.
func TestSecureChannelRolling(t *testing.T) {
	seed := msnrpc.NETLOGON_CREDENTIAL{0x7e, 0x0a, 0xa0, 0x52, 0x63, 0xd9, 0xe4, 0xee}
	timestamps := []uint32{0x1000, 0x1005, 0x100a, 0x2000}
	idx := 0
	sc := &SecureChannel{
		sessionKey:       scSessionKey,
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
		serverStored = addToCredential(serverStored, ts)
		if want := ComputeNetlogonCredentialAES(serverStored, scSessionKey); want != auth.Credential {
			t.Fatalf("ts %#x: client credential %x, server expected %x", ts, auth.Credential, want)
		}
		serverStored = addToCredential(serverStored, 1)
		resp := msnrpc.NETLOGON_AUTHENTICATOR{
			Credential: ComputeNetlogonCredentialAES(serverStored, scSessionKey),
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
		clientStoredCred: msnrpc.NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8},
		now:              func() time.Time { return time.Unix(0x3000, 0) },
	}
	sc.NextAuthenticator()
	if err := sc.VerifyResponseAuthenticator(msnrpc.NETLOGON_AUTHENTICATOR{}); err == nil {
		t.Fatal("VerifyResponseAuthenticator accepted an all-zero server authenticator")
	}
}

// scriptInvoker answers the two handshake RPCs from canned data, records the opnums seen and
// the credential the client sent, and can forge the server credential or fail ReqChallenge.
type scriptInvoker struct {
	serverChallenge msnrpc.NETLOGON_CREDENTIAL
	sessionKey      [16]byte
	reqStatus       uint32
	forgeServer     bool
	opnums          []uint16
	gotClientCred   msnrpc.NETLOGON_CREDENTIAL
	gotAccountName  ndr.WSTR
}

func (s *scriptInvoker) Invoke(in ndr.Call, out any) error {
	s.opnums = append(s.opnums, in.Opnum())
	switch o := out.(type) {
	case *netrServerReqChallengeResponse:
		o.ServerChallenge = s.serverChallenge
		o.Status = ndr.DWORD(s.reqStatus)
	case *netrServerAuthenticate3Response:
		req := in.(*netrServerAuthenticate3Request)
		s.gotClientCred = req.ClientCredential
		s.gotAccountName = req.AccountName
		cred := ComputeNetlogonCredentialAES(s.serverChallenge, s.sessionKey)
		if s.forgeServer {
			cred[0] ^= 0xff
		}
		o.ServerCredential = cred
		o.NegotiateFlags = req.NegotiateFlags
		o.Status = ndr.DWORD(netlogon.StatusSuccess)
	default:
		return fmt.Errorf("scriptInvoker: unexpected response type %T", out)
	}
	return nil
}

// establishFixture builds a scripted invoker and config for a fixed client/server challenge
// and password, so the handshake is fully deterministic.
func establishFixture(forge bool, reqStatus uint32) (*scriptInvoker, SecureChannelConfig, msnrpc.NETLOGON_CREDENTIAL, [16]byte) {
	clientChallenge := msnrpc.NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8}
	serverChallenge := msnrpc.NETLOGON_CREDENTIAL{8, 7, 6, 5, 4, 3, 2, 1}
	const password = "Machine$Pass1"
	sk := ComputeSessionKeyAES(password, nil, clientChallenge, serverChallenge)
	inv := &scriptInvoker{serverChallenge: serverChallenge, sessionKey: sk, reqStatus: reqStatus, forgeServer: forge}
	cfg := SecureChannelConfig{
		PrimaryName:       "DC01",
		ComputerName:      "WS01",
		AccountName:       "WS01$",
		SecureChannelType: msnrpc.WorkstationSecureChannel,
		Password:          password,
		Rand:              bytes.NewReader(clientChallenge[:]),
	}
	return inv, cfg, clientChallenge, sk
}

// TestEstablishHappyPath verifies the handshake sequence, the derived session key, the
// seeded stored credential, and the credential the client sent in NetrServerAuthenticate3.
func TestEstablishHappyPath(t *testing.T) {
	inv, cfg, clientChallenge, sk := establishFixture(false, netlogon.StatusSuccess)

	sc, err := Establish(inv, cfg)
	if err != nil {
		t.Fatalf("Establish: %v", err)
	}
	if inv.opnums[0] != netlogon.OpnumNetrServerReqChallenge || inv.opnums[1] != netlogon.OpnumNetrServerAuthenticate3 {
		t.Fatalf("opnum sequence = %v, want [4 26]", inv.opnums)
	}
	if sc.SessionKey() != sk {
		t.Errorf("session key = %x, want %x", sc.SessionKey(), sk)
	}
	// The stored credential is seeded to the client credential = Compute(clientChallenge).
	wantSeed := ComputeNetlogonCredentialAES(clientChallenge, sk)
	if sc.clientStoredCred != wantSeed {
		t.Errorf("seeded stored credential = %x, want %x", sc.clientStoredCred, wantSeed)
	}
	if inv.gotClientCred != wantSeed {
		t.Errorf("client credential sent = %x, want %x", inv.gotClientCred, wantSeed)
	}
	if inv.gotAccountName != ndr.WSTR("WS01$") {
		t.Errorf("account name sent = %q, want %q", inv.gotAccountName, "WS01$")
	}
	// AES must be negotiated regardless of the caller's flags (cfg left NegotiateFlags 0).
	if sc.NegotiateFlags()&netlogon.NegotiateAES == 0 {
		t.Errorf("negotiate flags %#x missing NegotiateAES", sc.NegotiateFlags())
	}
}

// TestEstablishRejectsForgedServerCredential confirms Establish fails when the server does
// not prove knowledge of the shared secret.
func TestEstablishRejectsForgedServerCredential(t *testing.T) {
	inv, cfg, _, _ := establishFixture(true, netlogon.StatusSuccess)
	if _, err := Establish(inv, cfg); err == nil {
		t.Fatal("Establish accepted a forged server credential")
	}
}

// TestEstablishReqChallengeStatusError confirms a non-success ReqChallenge status aborts the
// handshake before NetrServerAuthenticate3 is called.
func TestEstablishReqChallengeStatusError(t *testing.T) {
	inv, cfg, _, _ := establishFixture(false, netlogon.StatusAccessDenied)
	if _, err := Establish(inv, cfg); err == nil {
		t.Fatal("Establish ignored a failed NetrServerReqChallenge status")
	}
	if len(inv.opnums) != 1 || inv.opnums[0] != netlogon.OpnumNetrServerReqChallenge {
		t.Fatalf("opnums = %v, want just [4] (no Authenticate3 after failure)", inv.opnums)
	}
}
