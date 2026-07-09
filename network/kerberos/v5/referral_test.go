package kerberos

import (
	"encoding/asn1"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// srvName is a small helper to build a service PrincipalName.
func srvName(parts ...string) messages.PrincipalName {
	return messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: parts}
}

// TestReferralRealmFromSName covers the core referral-TGT detection: a returned
// ticket whose server name is krbtgt/OTHER-REALM (differing from the requested
// service) signals a cross-realm referral toward OTHER-REALM.
func TestReferralRealmFromSName(t *testing.T) {
	requested := srvName("cifs", "host.child.corp.local")

	tests := []struct {
		name      string
		ticket    messages.PrincipalName
		wantRealm string
		wantRef   bool
	}{
		{
			name:      "referral TGT to child realm",
			ticket:    srvName("krbtgt", "CHILD.CORP.LOCAL"),
			wantRealm: "CHILD.CORP.LOCAL",
			wantRef:   true,
		},
		{
			name:      "krbtgt case-insensitive service label",
			ticket:    srvName("KrbTgt", "CHILD.CORP.LOCAL"),
			wantRealm: "CHILD.CORP.LOCAL",
			wantRef:   true,
		},
		{
			name:    "actual service ticket is not a referral",
			ticket:  srvName("cifs", "host.child.corp.local"),
			wantRef: false,
		},
		{
			name:    "single-component name is not a referral",
			ticket:  messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt"}},
			wantRef: false,
		},
		{
			name:    "explicitly requested krbtgt is not a referral",
			ticket:  srvName("krbtgt", "CORP.LOCAL"),
			wantRef: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := requested
			if tt.name == "explicitly requested krbtgt is not a referral" {
				req = srvName("krbtgt", "CORP.LOCAL")
			}
			realm, isRef := referralRealmFromSName(tt.ticket, req)
			if isRef != tt.wantRef {
				t.Fatalf("isReferral = %v, want %v", isRef, tt.wantRef)
			}
			if isRef && realm != tt.wantRealm {
				t.Errorf("realm = %q, want %q", realm, tt.wantRealm)
			}
		})
	}
}

// TestReferralTargetRealmFromSyntheticTGSRep feeds a synthetic referral TGS-REP
// (server name krbtgt/CHILD.CORP.LOCAL) and asserts the chase re-targets the
// child realm.
func TestReferralTargetRealmFromSyntheticTGSRep(t *testing.T) {
	rep := &messages.TGSRep{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSRep,
		CRealm:  "CORP.LOCAL",
		Ticket: messages.Ticket{
			TktVno: messages.KerberosV5,
			Realm:  "CORP.LOCAL", // issuing realm
			SName:  srvName("krbtgt", "CHILD.CORP.LOCAL"),
		},
	}
	requested := srvName("cifs", "host.child.corp.local")

	realm, isRef := referralTargetRealm(rep, requested)
	if !isRef {
		t.Fatalf("expected referral to be detected")
	}
	if realm != "CHILD.CORP.LOCAL" {
		t.Errorf("next realm = %q, want CHILD.CORP.LOCAL", realm)
	}
}

// TestReferralTargetRealmPrefersSvrReferralInfo verifies that when the KDC also
// supplies a PA-SVR-REFERRAL-INFO element, its referred-realm is used as the
// authoritative next-realm hint.
func TestReferralTargetRealmPrefersSvrReferralInfo(t *testing.T) {
	rep := &messages.TGSRep{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSRep,
		Ticket: messages.Ticket{
			SName: srvName("krbtgt", "INTERMEDIATE.CORP.LOCAL"),
		},
		PAData: []messages.PAData{
			{PADataType: messages.PASvrReferralInfo, PADataValue: mustSvrReferralData(t, "TRUE.CORP.LOCAL")},
		},
	}
	realm, isRef := referralTargetRealm(rep, srvName("cifs", "host.true.corp.local"))
	if !isRef {
		t.Fatalf("expected referral to be detected")
	}
	if realm != "TRUE.CORP.LOCAL" {
		t.Errorf("next realm = %q, want TRUE.CORP.LOCAL (from PA-SVR-REFERRAL-INFO)", realm)
	}
}

// TestSvrReferralRealm parses a PA-SVR-REFERRAL-DATA carrying only referred-realm
// (the form Windows KDCs actually emit).
func TestSvrReferralRealm(t *testing.T) {
	pa := []messages.PAData{
		{PADataType: messages.PAETypeInfo2, PADataValue: []byte{0x30, 0x00}},
		{PADataType: messages.PASvrReferralInfo, PADataValue: mustSvrReferralData(t, "child.corp.local")},
	}
	realm, ok := svrReferralRealm(pa)
	if !ok {
		t.Fatalf("expected PA-SVR-REFERRAL-INFO to be found")
	}
	if realm != "CHILD.CORP.LOCAL" {
		t.Errorf("realm = %q, want CHILD.CORP.LOCAL (uppercased)", realm)
	}

	if _, ok := svrReferralRealm(pa[:1]); ok {
		t.Errorf("expected no referral info when the element is absent")
	}
}

// TestResolveKDCForRealm exercises the resolution precedence: explicit overrides,
// the client's home realm, and a custom resolver — none hardcoded.
func TestResolveKDCForRealm(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1")
	c.WithRealmKDC("child.corp.local", "10.0.1.1")
	c.WithKDCResolver(func(realm string) (string, error) {
		return "resolver-" + realm, nil
	})

	// Explicit override (case-insensitive realm key).
	if host, err := c.resolveKDCForRealm("CHILD.CORP.LOCAL"); err != nil || host != "10.0.1.1" {
		t.Errorf("override: got (%q, %v), want (10.0.1.1, nil)", host, err)
	}
	// Home realm resolves to the configured KDC.
	if host, err := c.resolveKDCForRealm("CORP.LOCAL"); err != nil || host != "10.0.0.1" {
		t.Errorf("home realm: got (%q, %v), want (10.0.0.1, nil)", host, err)
	}
	// Unknown realm falls through to the custom resolver.
	if host, err := c.resolveKDCForRealm("OTHER.LOCAL"); err != nil || host != "resolver-OTHER.LOCAL" {
		t.Errorf("resolver: got (%q, %v), want (resolver-OTHER.LOCAL, nil)", host, err)
	}
}

// TestPrincipalNameEqualFold checks the case-insensitive component comparison.
func TestPrincipalNameEqualFold(t *testing.T) {
	a := srvName("krbtgt", "CORP.LOCAL")
	b := srvName("KRBTGT", "corp.local")
	if !principalNameEqualFold(a, b) {
		t.Errorf("expected case-insensitive equality")
	}
	if principalNameEqualFold(a, srvName("krbtgt", "OTHER.LOCAL")) {
		t.Errorf("expected inequality for differing realm")
	}
	if principalNameEqualFold(a, srvName("krbtgt")) {
		t.Errorf("expected inequality for differing length")
	}
}

// mustSvrReferralData builds a DER PA-SVR-REFERRAL-DATA containing only the
// referred-realm [0] (a GeneralString), matching what a KDC sends.
func mustSvrReferralData(t *testing.T, realm string) []byte {
	t.Helper()
	realmElem, err := asn1.Marshal(messages.ExplicitGeneralString(0, realm))
	if err != nil {
		t.Fatalf("marshal referred-realm: %v", err)
	}
	seq, err := asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSequence,
		IsCompound: true,
		Bytes:      realmElem,
	})
	if err != nil {
		t.Fatalf("marshal SEQUENCE: %v", err)
	}
	return seq
}

// TestKRBErrorWrongRealmCRealm confirms a KRB-ERROR round-trips its CRealm field,
// which the referral chase reads to recover from KDC_ERR_WRONG_REALM.
func TestKRBErrorWrongRealmCRealm(t *testing.T) {
	orig := &messages.KRBError{
		ErrorCode: messages.ErrWrongRealm,
		CRealm:    "TRUE.CORP.LOCAL",
		Realm:     "CORP.LOCAL",
		SName:     srvName("krbtgt", "CORP.LOCAL"),
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("KRBError.Marshal: %v", err)
	}
	var decoded messages.KRBError
	if _, err := decoded.Unmarshal(wire); err != nil {
		t.Fatalf("KRBError.Unmarshal: %v", err)
	}
	if decoded.ErrorCode != messages.ErrWrongRealm {
		t.Errorf("ErrorCode = %d, want %d", decoded.ErrorCode, messages.ErrWrongRealm)
	}
	if decoded.CRealm != "TRUE.CORP.LOCAL" {
		t.Errorf("CRealm = %q, want TRUE.CORP.LOCAL", decoded.CRealm)
	}
}
