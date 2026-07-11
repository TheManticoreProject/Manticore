package kerberos

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
)

// TestForgeDiamondInjectsGroupAndResigns builds a synthetic "KDC-issued" TGT
// (a golden ticket doubles as a stand-in: it is encrypted with the krbtgt key at
// ticket key usage 2 and carries a krbtgt-signed PAC), then diamonds it: decrypts
// with the krbtgt key, injects Domain Admins, re-signs and re-encrypts. It then
// decrypts the diamond ticket and asserts the injected group is present, the
// original group and identity are preserved, and both PAC signatures verify.
func TestForgeDiamondInjectsGroupAndResigns(t *testing.T) {
	krbtgt := bytes.Repeat([]byte{0x42}, 32) // krbtgt AES256 key
	etype := messages.ETypeAES256CTSHMACSHA196

	base, err := ForgeGolden(ForgeOptions{
		Realm:           testRealm,
		Username:        "lowpriv",
		DomainSID:       testDomainSID,
		UserRID:         1105,
		PrimaryGroupRID: 513,
		GroupRIDs:       []uint32{513}, // only Domain Users
		LogonDomainName: "CORP",
		LogonServer:     "DC01",
		Key:             krbtgt,
		KeyEType:        etype,
		StartTime:       time.Unix(1700000000, 0),
	})
	if err != nil {
		t.Fatalf("build synthetic KDC-issued TGT: %v", err)
	}

	diamond, err := forgeDiamondFromTicket(base.Ticket, krbtgt, PACModifications{
		AddGroupRIDs: []uint32{512}, // Domain Admins
	})
	if err != nil {
		t.Fatalf("forgeDiamondFromTicket: %v", err)
	}

	// The client identity and service name must be preserved (still the low-priv
	// user's genuine TGT for krbtgt/REALM).
	enc, pacBytes := extractPAC(t, diamond.Ticket, krbtgt, etype)
	if len(enc.CName.NameString) != 1 || enc.CName.NameString[0] != "lowpriv" {
		t.Errorf("diamond CName = %+v, want lowpriv (identity must be preserved)", enc.CName)
	}
	if diamond.Ticket.SName.NameString[0] != "krbtgt" {
		t.Errorf("diamond SName = %+v, want krbtgt/REALM", diamond.Ticket.SName)
	}

	p, err := pac.Parse(pacBytes)
	if err != nil {
		t.Fatalf("parse diamond PAC: %v", err)
	}
	if err := p.VerifyServerSignature(krbtgt); err != nil {
		t.Errorf("diamond PAC server signature: %v", err)
	}
	if err := p.VerifyKDCSignature(krbtgt); err != nil {
		t.Errorf("diamond PAC KDC signature: %v", err)
	}

	info, err := p.LogonInfo()
	if err != nil {
		t.Fatalf("decode diamond PAC logon info: %v", err)
	}
	if info.UserRID() != 1105 {
		t.Errorf("diamond UserRID = %d, want 1105 (unchanged)", info.UserRID())
	}
	if !hasRID(info.GroupRIDs(), 512) {
		t.Errorf("injected group 512 (Domain Admins) missing; groups=%v", info.GroupRIDs())
	}
	if !hasRID(info.GroupRIDs(), 513) {
		t.Errorf("original group 513 (Domain Users) lost; groups=%v", info.GroupRIDs())
	}

	// The diamond ticket must round-trip through the pass-the-ticket import path.
	kirbi, err := diamond.KirbiBytes()
	if err != nil {
		t.Fatalf("KirbiBytes: %v", err)
	}
	c := NewClient("", "", "")
	if err := c.LoadTGTFromKirbiBytes(kirbi); err != nil {
		t.Fatalf("LoadTGTFromKirbiBytes: %v", err)
	}
	if c.Username() != "lowpriv" || c.Realm() != testRealm {
		t.Errorf("imported identity = %q@%q, want lowpriv@%s", c.Username(), c.Realm(), testRealm)
	}
}

// TestForgeDiamondNoModsRoundTrips confirms that with no modifications the PAC
// logon-info round-trips faithfully through decode/re-encode and both signatures
// still verify — the fidelity the live path relies on.
func TestForgeDiamondNoModsRoundTrips(t *testing.T) {
	krbtgt := bytes.Repeat([]byte{0x7E}, 16) // RC4 krbtgt key
	etype := messages.ETypeRC4HMAC

	base, err := ForgeGolden(ForgeOptions{
		Realm: testRealm, Username: "svc", DomainSID: testDomainSID, UserRID: 1120,
		GroupRIDs: []uint32{513, 512, 519}, LogonDomainName: "CORP",
		Key: krbtgt, KeyEType: etype,
	})
	if err != nil {
		t.Fatalf("ForgeGolden: %v", err)
	}
	_, origPACBytes := extractPAC(t, base.Ticket, krbtgt, etype)
	origPAC, _ := pac.Parse(origPACBytes)
	origInfo, _ := origPAC.LogonInfo()

	diamond, err := forgeDiamondFromTicket(base.Ticket, krbtgt, PACModifications{})
	if err != nil {
		t.Fatalf("forgeDiamondFromTicket: %v", err)
	}
	_, pacBytes := extractPAC(t, diamond.Ticket, krbtgt, etype)
	p, _ := pac.Parse(pacBytes)
	if err := p.VerifyServerSignature(krbtgt); err != nil {
		t.Errorf("server signature: %v", err)
	}
	if err := p.VerifyKDCSignature(krbtgt); err != nil {
		t.Errorf("KDC signature: %v", err)
	}
	info, _ := p.LogonInfo()
	if info.UserRID() != origInfo.UserRID() || info.UserName() != origInfo.UserName() {
		t.Errorf("identity changed by no-op diamond: got %q/%d want %q/%d",
			info.UserName(), info.UserRID(), origInfo.UserName(), origInfo.UserRID())
	}
	if len(info.GroupRIDs()) != len(origInfo.GroupRIDs()) {
		t.Errorf("group set changed by no-op diamond: got %v want %v", info.GroupRIDs(), origInfo.GroupRIDs())
	}
}

// TestForgeSapphireGraftsPACIntact builds a synthetic S4U2Self+U2U reply ticket
// carrying a genuine privileged PAC (encrypted, like an ENC-TKT-IN-SKEY reply,
// under a TGT session key), then extracts and grafts it. It asserts the extracted
// PAC bytes are the genuine buffer verbatim, that the graft leaves every PAC
// identity field unchanged, and that the grafted TGT is re-signed with the krbtgt
// key.
func TestForgeSapphireGraftsPACIntact(t *testing.T) {
	svcKey := bytes.Repeat([]byte{0x11}, 16)  // service account key (RC4)
	krbtgt := bytes.Repeat([]byte{0x99}, 32)  // domain krbtgt key (AES256)
	sessKey := bytes.Repeat([]byte{0x33}, 16) // TGT session key the reply is sealed to (RC4)

	// A genuine privileged PAC, signed as the KDC would (service key as the
	// "server", here also standing in for the counter-signer).
	genuineInfo, err := buildLogonInfo(&ForgeOptions{
		Realm: testRealm, Username: "Administrator", DomainSID: testDomainSID,
		UserRID: 500, PrimaryGroupRID: 513, GroupRIDs: []uint32{513, 512, 519, 518, 520},
		LogonDomainName: "CORP", UserAccountControl: defaultUserAccountControl,
	}, time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("buildLogonInfo: %v", err)
	}
	genuinePACObj, err := pac.Forge(genuineInfo, "Administrator", time.Unix(1700000000, 0), messages.ETypeRC4HMAC)
	if err != nil {
		t.Fatalf("pac.Forge: %v", err)
	}
	genuinePAC, err := genuinePACObj.Sign(svcKey, svcKey)
	if err != nil {
		t.Fatalf("sign genuine PAC: %v", err)
	}

	// Wrap it in a ticket enc-part encrypted under the TGT session key (usage 2),
	// mirroring the ENC-TKT-IN-SKEY reply the sapphire harvest decrypts.
	replyTicket := syntheticTicketWithPAC(t, genuinePAC, "Administrator", testRealm, sessKey, messages.ETypeRC4HMAC)

	// Extract (as harvestPACViaS4USelfU2U does) and assert the PAC came out intact.
	extracted, cname, crealm, err := extractPACFromTicket(replyTicket, messages.ETypeRC4HMAC, sessKey)
	if err != nil {
		t.Fatalf("extractPACFromTicket: %v", err)
	}
	if !bytes.Equal(extracted, genuinePAC) {
		t.Fatal("extracted PAC bytes differ from the genuine PAC (extraction not intact)")
	}
	if len(cname.NameString) != 1 || cname.NameString[0] != "Administrator" {
		t.Errorf("harvested client = %+v, want Administrator", cname)
	}

	// Graft into a sapphire TGT re-signed with the krbtgt key.
	sapphire, err := graftPACIntoTGT(extracted, cname, crealm, SapphireOptions{
		ImpersonateUser: "Administrator",
		Key:             krbtgt,
		KeyEType:        messages.ETypeAES256CTSHMACSHA196,
	})
	if err != nil {
		t.Fatalf("graftPACIntoTGT: %v", err)
	}

	if sapphire.Ticket.SName.NameString[0] != "krbtgt" || sapphire.Ticket.SName.NameString[1] != testRealm {
		t.Errorf("sapphire SName = %+v, want krbtgt/%s", sapphire.Ticket.SName, testRealm)
	}
	enc, pacBytes := extractPAC(t, sapphire.Ticket, krbtgt, messages.ETypeAES256CTSHMACSHA196)
	if len(enc.CName.NameString) != 1 || enc.CName.NameString[0] != "Administrator" {
		t.Errorf("sapphire CName = %+v, want Administrator", enc.CName)
	}

	p, err := pac.Parse(pacBytes)
	if err != nil {
		t.Fatalf("parse grafted PAC: %v", err)
	}
	// Re-signed with the krbtgt key (both signatures), not the original service key.
	if err := p.VerifyServerSignature(krbtgt); err != nil {
		t.Errorf("grafted PAC server signature (krbtgt): %v", err)
	}
	if err := p.VerifyKDCSignature(krbtgt); err != nil {
		t.Errorf("grafted PAC KDC signature (krbtgt): %v", err)
	}

	// Every identity field must be the genuine one — nothing fabricated.
	grafted, err := p.LogonInfo()
	if err != nil {
		t.Fatalf("decode grafted PAC logon info: %v", err)
	}
	if grafted.UserName() != "Administrator" || grafted.UserRID() != 500 {
		t.Errorf("grafted identity = %q/%d, want Administrator/500", grafted.UserName(), grafted.UserRID())
	}
	if !hasRID(grafted.GroupRIDs(), 512) || !hasRID(grafted.GroupRIDs(), 519) {
		t.Errorf("grafted PAC lost genuine privileged groups: %v", grafted.GroupRIDs())
	}
	if len(grafted.GroupRIDs()) != len(genuineInfo.GroupRIDs()) {
		t.Errorf("grafted group count = %d, want %d (must be unmodified)", len(grafted.GroupRIDs()), len(genuineInfo.GroupRIDs()))
	}
}

// TestSapphireRequestShape checks the harvest TGS-REQ carries the U2U markers:
// ENC-TKT-IN-SKEY, an additional ticket, and a PA-FOR-USER element.
func TestSapphireRequestShape(t *testing.T) {
	c := fakeTGTClient(t)
	req, err := c.buildSapphireTGSReq("Administrator", "", 4242)
	if err != nil {
		t.Fatalf("buildSapphireTGSReq: %v", err)
	}
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var app asn1.RawValue
	if _, err := asn1.Unmarshal(wire, &app); err != nil {
		t.Fatal(err)
	}
	bodyElem := seqChild(t, app.Bytes, asn1.ClassContextSpecific, 4)

	optElem := seqChild(t, bodyElem.Bytes, asn1.ClassContextSpecific, 0)
	var flags asn1.BitString
	if _, err := asn1.Unmarshal(optElem.Bytes, &flags); err != nil {
		t.Fatalf("parse KDCOptions: %v", err)
	}
	if flags.At(kdcOptionEncTktInSKey) != 1 {
		t.Errorf("ENC-TKT-IN-SKEY (bit 28) not set; bytes=% X", flags.Bytes)
	}

	// [11] additional-tickets must be present (the client's own TGT).
	if !hasContextTag(bodyElem.Bytes, 11) {
		t.Error("no additional-tickets [11] in sapphire TGS-REQ body")
	}

	// A PA-FOR-USER PAData element (type 129) must be present.
	if !hasPAType(t, wire, iana.PAForUser) {
		t.Error("no PA-FOR-USER element in sapphire TGS-REQ")
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func hasRID(rids []uint32, want uint32) bool {
	for _, r := range rids {
		if r == want {
			return true
		}
	}
	return false
}

// syntheticTicketWithPAC builds a ticket whose enc-part (encrypted under key at
// ticket key usage 2) carries the given PAC for cname@crealm, modeling a
// KDC-issued reply ticket.
func syntheticTicketWithPAC(t *testing.T, pacBytes []byte, username, realm string, key []byte, etype int) messages.Ticket {
	t.Helper()
	authData, err := wrapPACInAuthData(pacBytes)
	if err != nil {
		t.Fatalf("wrapPACInAuthData: %v", err)
	}
	enc := messages.EncTicketPart{
		Flags:             messages.NewKerberosFlags(messages.TicketFlagForwardable),
		Key:               messages.EncryptionKey{KeyType: etype, KeyValue: key},
		CRealm:            realm,
		CName:             messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{username}},
		Transited:         messages.TransitedEncoding{TRType: 0, Contents: []byte{}},
		AuthTime:          time.Unix(1700000000, 0).UTC(),
		StartTime:         time.Unix(1700000000, 0).UTC(),
		EndTime:           time.Unix(1700003600, 0).UTC(),
		RenewTill:         time.Unix(1700007200, 0).UTC(),
		AuthorizationData: authData,
	}
	encBytes, err := enc.Marshal()
	if err != nil {
		t.Fatalf("marshal EncTicketPart: %v", err)
	}
	cipher, err := kerbcrypto.Encrypt(etype, key, kerbcrypto.KeyUsageKDCRepTicket, encBytes)
	if err != nil {
		t.Fatalf("encrypt EncTicketPart: %v", err)
	}
	return messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   realm,
		SName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{username}},
		EncPart: messages.EncryptedData{EType: etype, Cipher: cipher},
	}
}

// hasContextTag reports whether the SEQUENCE (given as its TLV) contains a
// context-specific element with the given tag.
func hasContextTag(seqTLV []byte, tag int) bool {
	var seq asn1.RawValue
	if _, err := asn1.Unmarshal(seqTLV, &seq); err != nil {
		return false
	}
	rest := seq.Bytes
	for len(rest) > 0 {
		var e asn1.RawValue
		r, err := asn1.Unmarshal(rest, &e)
		if err != nil {
			return false
		}
		if e.Class == asn1.ClassContextSpecific && e.Tag == tag {
			return true
		}
		rest = r
	}
	return false
}

// hasPAType reports whether the TGS-REQ wire carries a PAData element of padType.
func hasPAType(t *testing.T, wire []byte, padType int) bool {
	t.Helper()
	var req messages.TGSReq
	if _, err := req.Unmarshal(wire); err != nil {
		t.Fatalf("re-parse TGS-REQ: %v", err)
	}
	for _, pa := range req.PAData {
		if pa.PADataType == padType {
			return true
		}
	}
	return false
}
