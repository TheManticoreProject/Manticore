package gssapi

import (
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// TestInitSecContextAppliesClockOffset checks that the AP-REQ authenticator and
// the third-leg AP-REP are stamped with the local clock plus
// InitOptions.ClockOffset, so a service whose clock is hours away from ours,
// and which the Kerberos client has already learned to match at the KDC,
// accepts the AP-REQ instead of failing it with KRB_AP_ERR_SKEW.
func TestInitSecContextAppliesClockOffset(t *testing.T) {
	const etype = iana.ETypeAES256CTSHMACSHA196
	const offset = -12*time.Hour - 30*time.Minute
	serverNow := time.Now().UTC().Add(offset)

	serviceKey := randKey(t, 32)
	sessionKey := randKey(t, 32)
	client := messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"alice"}}
	ticketRaw := serviceTicketCustom(t, etype, serviceKey, sessionKey, client, "CORP.LOCAL", messages.EncTicketPart{
		AuthTime: serverNow.Add(-time.Minute),
		EndTime:  serverNow.Add(8 * time.Hour),
	})
	initOpts := InitOptions{
		TicketRaw: ticketRaw, SessionKey: sessionKey, SessionEType: etype,
		ClientName: client, ClientRealm: "CORP.LOCAL",
		Flags: GSSIntegFlag | GSSConfFlag, Mutual: true,
	}
	acceptOpts := serviceAcceptOptions(AcceptOptions{
		Keys:                   []ServiceKey{{EType: etype, Key: serviceKey}},
		Now:                    serverNow,
		ReplayCache:            NewReplayCache(),
		DisableReplayDetection: true,
	})

	// Without the offset the authenticator is 12.5 hours off the server clock.
	token, _, err := InitSecContext(initOpts)
	if err != nil {
		t.Fatalf("InitSecContext: %v", err)
	}
	if _, _, err := AcceptSecContext(token, acceptOpts); err == nil {
		t.Fatal("AcceptSecContext accepted an authenticator 12.5 hours off the server clock")
	}

	initOpts.ClockOffset = offset
	token, ictx, err := InitSecContext(initOpts)
	if err != nil {
		t.Fatalf("InitSecContext with ClockOffset: %v", err)
	}
	if _, _, err := AcceptSecContext(token, acceptOpts); err != nil {
		t.Fatalf("AcceptSecContext with ClockOffset: %v", err)
	}

	apRepRaw, err := ictx.MakeAPRep()
	if err != nil {
		t.Fatalf("MakeAPRep: %v", err)
	}
	var apRep messages.APRep
	if _, err := apRep.Unmarshal(apRepRaw); err != nil {
		t.Fatalf("parse AP-REP: %v", err)
	}
	plain, err := kerbcrypto.Decrypt(etype, sessionKey, kerbcrypto.KeyUsageAPRepEncPart, apRep.EncPart.Cipher)
	if err != nil {
		t.Fatalf("decrypt AP-REP: %v", err)
	}
	var enc messages.EncAPRepPart
	if _, err := enc.Unmarshal(plain); err != nil {
		t.Fatalf("parse EncAPRepPart: %v", err)
	}
	if d := enc.CTime.Sub(serverNow); d < -time.Minute || d > time.Minute {
		t.Errorf("third-leg AP-REP ctime is %v off the server clock, want within a minute", d)
	}
}
