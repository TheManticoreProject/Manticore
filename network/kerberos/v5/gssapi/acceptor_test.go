package gssapi

import (
	"bytes"
	"crypto/rand"
	"encoding/asn1"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/keytab"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// serviceTicket builds a synthetic service ticket for cifs/host@REALM whose
// enc-part is sealed under serviceKey (key usage 2), carrying sessionKey as the
// ticket session key and issued to clientName@clientRealm. When pacBytes is
// non-nil it is wrapped as an AD-IF-RELEVANT → AD-WIN2K-PAC element, exactly as a
// KDC would. It returns the raw APPLICATION[1] ticket bytes.
func serviceTicket(t *testing.T, etype int, serviceKey, sessionKey []byte, clientName messages.PrincipalName, clientRealm string, pacBytes []byte) []byte {
	t.Helper()
	now := time.Now().UTC()
	var ad []messages.AuthorizationData
	if pacBytes != nil {
		inner, err := asn1.Marshal([]messages.AuthorizationData{{ADType: adTypeWin2KPAC, ADData: pacBytes}})
		if err != nil {
			t.Fatalf("marshal inner AD: %v", err)
		}
		ad = []messages.AuthorizationData{{ADType: adTypeIfRelevant, ADData: inner}}
	}
	encPart := messages.EncTicketPart{
		Flags:             asn1.BitString{Bytes: []byte{0x40, 0, 0, 0}, BitLength: 32},
		Key:               messages.EncryptionKey{KeyType: etype, KeyValue: sessionKey},
		CRealm:            clientRealm,
		CName:             clientName,
		AuthTime:          now,
		EndTime:           now.Add(8 * time.Hour),
		AuthorizationData: ad,
	}
	plain, err := encPart.Marshal()
	if err != nil {
		t.Fatalf("marshal EncTicketPart: %v", err)
	}
	cipher, err := kerbcrypto.Encrypt(etype, serviceKey, kerbcrypto.KeyUsageKDCRepTicket, plain)
	if err != nil {
		t.Fatalf("encrypt ticket enc-part: %v", err)
	}
	tkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   clientRealm,
		SName:   messages.PrincipalName{NameType: iana.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}},
		EncPart: messages.EncryptedData{EType: etype, Cipher: cipher},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("marshal ticket: %v", err)
	}
	return raw
}

func randKey(t *testing.T, n int) []byte {
	t.Helper()
	k := make([]byte, n)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	return k
}

// exerciseBothDirections drives per-message MIC and Wrap tokens in both
// directions across an established initiator/acceptor pair.
func exerciseBothDirections(t *testing.T, ictx, actx *SecContext) {
	t.Helper()

	// Initiator -> acceptor MIC.
	msg := []byte("hello from the initiator")
	mic, err := ictx.MakeMIC(msg)
	if err != nil {
		t.Fatalf("initiator MakeMIC: %v", err)
	}
	if err := actx.VerifyMIC(msg, mic); err != nil {
		t.Fatalf("acceptor VerifyMIC (initiator MIC): %v", err)
	}

	// Acceptor -> initiator MIC.
	msg2 := []byte("hello from the acceptor")
	mic2, err := actx.MakeMIC(msg2)
	if err != nil {
		t.Fatalf("acceptor MakeMIC: %v", err)
	}
	if err := ictx.VerifyMIC(msg2, mic2); err != nil {
		t.Fatalf("initiator VerifyMIC (acceptor MIC): %v", err)
	}

	// Initiator -> acceptor Wrap (sealed).
	secret := []byte("confidential initiator payload")
	wrapped, err := ictx.Wrap(secret, true)
	if err != nil {
		t.Fatalf("initiator Wrap: %v", err)
	}
	got, sealed, err := actx.Unwrap(wrapped)
	if err != nil {
		t.Fatalf("acceptor Unwrap: %v", err)
	}
	if !sealed || !bytes.Equal(got, secret) {
		t.Fatalf("acceptor Unwrap = %q sealed=%v, want %q sealed=true", got, sealed, secret)
	}

	// Acceptor -> initiator Wrap (sealed).
	secret2 := []byte("confidential acceptor payload")
	wrapped2, err := actx.Wrap(secret2, true)
	if err != nil {
		t.Fatalf("acceptor Wrap: %v", err)
	}
	got2, sealed2, err := ictx.Unwrap(wrapped2)
	if err != nil {
		t.Fatalf("initiator Unwrap: %v", err)
	}
	if !sealed2 || !bytes.Equal(got2, secret2) {
		t.Fatalf("initiator Unwrap = %q sealed=%v, want %q sealed=true", got2, sealed2, secret2)
	}
}

// loopback runs InitSecContext -> AcceptSecContext for the given options and
// returns both contexts and the AP-REP output token.
func loopback(t *testing.T, etype int, initOpts InitOptions, acceptOpts AcceptOptions) (ictx, actx *SecContext, apRep []byte) {
	t.Helper()
	token, ictx, err := InitSecContext(initOpts)
	if err != nil {
		t.Fatalf("InitSecContext: %v", err)
	}
	apRep, actx, err = AcceptSecContext(token, acceptOpts)
	if err != nil {
		t.Fatalf("AcceptSecContext: %v", err)
	}
	return ictx, actx, apRep
}

func TestAcceptSecContextLoopbackMutual(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"alice"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	ictx, actx, apRep := loopback(t, etype,
		InitOptions{
			TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
			ClientName: client, ClientRealm: "CORP.LOCAL",
			Flags: GSSIntegFlag | GSSConfFlag, Mutual: true,
		},
		AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}},
	)

	// The acceptor recovered the client identity and session key.
	name, realm := actx.ClientPrincipal()
	if realm != "CORP.LOCAL" || len(name.NameString) != 1 || name.NameString[0] != "alice" {
		t.Errorf("ClientPrincipal = %v@%s, want alice@CORP.LOCAL", name.NameString, realm)
	}
	if !bytes.Equal(actx.SessionKey, sessionKey) {
		t.Error("acceptor session key does not match the ticket session key")
	}
	if actx.Authenticator() == nil {
		t.Error("acceptor did not retain the authenticator")
	}

	// Mutual authentication: the initiator accepts the acceptor's AP-REP.
	if len(apRep) == 0 {
		t.Fatal("expected an AP-REP output token for mutual authentication")
	}
	if err := ictx.AcceptAPRep(apRep); err != nil {
		t.Fatalf("initiator AcceptAPRep: %v", err)
	}

	exerciseBothDirections(t, ictx, actx)
}

func TestAcceptSecContextLoopbackViaKeytab(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"bob"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	kt := keytab.New()
	kt.Add(keytab.Principal{Realm: "CORP.LOCAL", Components: []string{"cifs", "host.corp.local"}}, etype, serviceKey, 3)

	ictx, actx, apRep := loopback(t, etype,
		InitOptions{
			TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
			ClientName: client, ClientRealm: "CORP.LOCAL", Mutual: true,
		},
		AcceptOptions{Keytab: kt},
	)
	if err := ictx.AcceptAPRep(apRep); err != nil {
		t.Fatalf("initiator AcceptAPRep: %v", err)
	}
	exerciseBothDirections(t, ictx, actx)
}

func TestAcceptSecContextRC4BothDirections(t *testing.T) {
	const etype = iana.ETypeRC4HMAC
	serviceKey := randKey(t, 16)
	sessionKey := randKey(t, 16)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"carol"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	ictx, actx, apRep := loopback(t, etype,
		InitOptions{
			TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
			ClientName: client, ClientRealm: "CORP.LOCAL", Mutual: true,
		},
		AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}},
	)
	if err := ictx.AcceptAPRep(apRep); err != nil {
		t.Fatalf("initiator AcceptAPRep: %v", err)
	}
	exerciseBothDirections(t, ictx, actx)
}

func TestAcceptSecContextMintSubkey(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"dan"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	ictx, actx, apRep := loopback(t, etype,
		InitOptions{
			TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
			ClientName: client, ClientRealm: "CORP.LOCAL", Mutual: true,
		},
		AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}, MintSubkey: true},
	)
	// The acceptor keys per-message tokens with its minted subkey; the initiator
	// must adopt it from the AP-REP before per-message tokens agree.
	if err := ictx.AcceptAPRep(apRep); err != nil {
		t.Fatalf("initiator AcceptAPRep: %v", err)
	}
	if len(actx.SubKey) == 0 || !bytes.Equal(actx.SubKey, ictx.SubKey) {
		t.Fatal("initiator did not adopt the acceptor subkey")
	}
	exerciseBothDirections(t, ictx, actx)
}

func TestAcceptSecContextAuthenticatorSubkey(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	subKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"erin"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	ictx, actx, apRep := loopback(t, etype,
		InitOptions{
			TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
			ClientName: client, ClientRealm: "CORP.LOCAL", Mutual: true,
			SubKey: subKey, SubKeyEType: etype,
		},
		AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}},
	)
	if !bytes.Equal(actx.SubKey, subKey) {
		t.Fatal("acceptor did not adopt the authenticator subkey")
	}
	if err := ictx.AcceptAPRep(apRep); err != nil {
		t.Fatalf("initiator AcceptAPRep: %v", err)
	}
	exerciseBothDirections(t, ictx, actx)
}

func TestAcceptSecContextNoMutualNoAPRep(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"frank"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL", Mutual: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	apRep, actx, err := AcceptSecContext(token, AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}})
	if err != nil {
		t.Fatalf("AcceptSecContext: %v", err)
	}
	if apRep != nil {
		t.Error("no AP-REP expected when mutual authentication was not requested")
	}
	if actx == nil {
		t.Fatal("expected a SecContext even without mutual authentication")
	}
}

func TestAcceptSecContextPACExtraction(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"grace"}}
	pacBytes := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03, 0x04}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", pacBytes)

	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, actx, err := AcceptSecContext(token, AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}})
	if err != nil {
		t.Fatalf("AcceptSecContext: %v", err)
	}
	if !actx.HasPAC() {
		t.Fatal("expected the acceptor to extract a PAC")
	}
	if !bytes.Equal(actx.PACBytes(), pacBytes) {
		t.Errorf("PACBytes = % X, want % X", actx.PACBytes(), pacBytes)
	}
}

func TestAcceptSecContextChannelBindings(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"heidi"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	cb := GSSChannelBindings([]byte("tls-server-end-point:abcd"))
	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL", ChannelBindings: cb,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Matching channel bindings pass.
	if _, _, err := AcceptSecContext(token, AcceptOptions{
		Keys: []ServiceKey{{EType: etype, Key: serviceKey}}, ChannelBindings: cb,
	}); err != nil {
		t.Fatalf("matching channel bindings rejected: %v", err)
	}

	// Mismatched channel bindings fail (tampered checksum Bnd field).
	if _, _, err := AcceptSecContext(token, AcceptOptions{
		Keys:            []ServiceKey{{EType: etype, Key: serviceKey}},
		ChannelBindings: GSSChannelBindings([]byte("tls-server-end-point:WRONG")),
	}); err == nil {
		t.Error("expected channel-binding mismatch to be rejected")
	}
}

func TestAcceptSecContextWrongServiceKey(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"ivan"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL",
	})
	if err != nil {
		t.Fatal(err)
	}
	wrong := randKey(t, 32)
	if _, _, err := AcceptSecContext(token, AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: wrong}}}); err == nil {
		t.Error("expected AcceptSecContext to fail with the wrong service key")
	}
}

func TestAcceptSecContextReplay(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"judy"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL",
	})
	if err != nil {
		t.Fatal(err)
	}
	rc := NewReplayCache()
	opts := AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}, ReplayCache: rc}

	if _, _, err := AcceptSecContext(token, opts); err != nil {
		t.Fatalf("first AcceptSecContext: %v", err)
	}
	// The same authenticator replayed against the shared cache must be rejected.
	if _, _, err := AcceptSecContext(token, opts); err == nil {
		t.Error("expected a replayed authenticator to be rejected")
	}
}

func TestAcceptSecContextClockSkew(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"kate"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL",
	})
	if err != nil {
		t.Fatal(err)
	}
	// The acceptor's clock is an hour ahead of the authenticator: skew rejected.
	skewed := AcceptOptions{
		Keys: []ServiceKey{{EType: etype, Key: serviceKey}},
		Now:  time.Now().UTC().Add(time.Hour),
	}
	if _, _, err := AcceptSecContext(token, skewed); err == nil {
		t.Error("expected an out-of-skew authenticator to be rejected")
	}
}

func TestAcceptSecContextTamperedAuthenticator(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"leo"}}
	ticketRaw := serviceTicket(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", nil)

	token, _, err := InitSecContext(InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Flip a byte late in the token (inside the encrypted authenticator): the
	// integrity check on the authenticator must fail.
	tampered := append([]byte(nil), token...)
	tampered[len(tampered)-5] ^= 0xff
	if _, _, err := AcceptSecContext(tampered, AcceptOptions{Keys: []ServiceKey{{EType: etype, Key: serviceKey}}}); err == nil {
		t.Error("expected a tampered authenticator to be rejected")
	}
}
