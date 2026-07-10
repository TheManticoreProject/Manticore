package messages

import (
	"bytes"
	"testing"
)

// TestKRBErrorRoundTrip marshals a KRB-ERROR (with the optional CRealm, EText
// and EData fields set) and confirms every field survives the round-trip.
func TestKRBErrorRoundTrip(t *testing.T) {
	edata := []byte{0x30, 0x03, 0x02, 0x01, 0x11}
	orig := &KRBError{
		STime:     tstTime,
		SUSec:     123456,
		ErrorCode: 25, // KDC_ERR_PREAUTH_REQUIRED
		CRealm:    "CORP.LOCAL",
		Realm:     "CORP.LOCAL",
		SName:     tgtSName(),
		EText:     "preauth required",
		EData:     edata,
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got KRBError
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.ErrorCode != 25 || got.SUSec != 123456 {
		t.Errorf("code/usec: %d/%d", got.ErrorCode, got.SUSec)
	}
	if !got.STime.Equal(tstTime) {
		t.Errorf("stime: got %v want %v", got.STime, tstTime)
	}
	if got.Realm != "CORP.LOCAL" || got.CRealm != "CORP.LOCAL" || got.SName.NameString[0] != "krbtgt" {
		t.Errorf("realms/sname not preserved: %+v", got)
	}
	if got.EText != "preauth required" {
		t.Errorf("etext: %q", got.EText)
	}
	if !bytes.Equal(got.EData, edata) {
		t.Errorf("edata: got %X want %X", got.EData, edata)
	}
}

// TestAuthenticatorRoundTrip marshals an Authenticator (with an optional checksum
// and sub-key) and confirms the identity, timestamp, checksum and sub-key round-trip.
func TestAuthenticatorRoundTrip(t *testing.T) {
	orig := &Authenticator{
		AVno:      KerberosV5,
		CRealm:    "CORP.LOCAL",
		CName:     cliName(),
		Cksum:     &Checksum{CKSumType: 16, Checksum: bytes.Repeat([]byte{0x7f}, 12)},
		CUSec:     654321,
		CTime:     tstTime,
		SubKey:    &EncryptionKey{KeyType: ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x22}, 32)},
		SeqNumber: 0x11223344,
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Authenticator
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.CRealm != "CORP.LOCAL" || got.CName.NameString[0] != "alice" || got.CUSec != 654321 {
		t.Errorf("identity not preserved: %+v", got)
	}
	if !got.CTime.Equal(tstTime) || got.SeqNumber != 0x11223344 {
		t.Errorf("ctime/seq: %v/%d", got.CTime, got.SeqNumber)
	}
	if got.Cksum == nil || got.Cksum.CKSumType != 16 || !bytes.Equal(got.Cksum.Checksum, orig.Cksum.Checksum) {
		t.Errorf("checksum not preserved: %+v", got.Cksum)
	}
	if got.SubKey == nil || got.SubKey.KeyType != ETypeAES256CTSHMACSHA196 ||
		!bytes.Equal(got.SubKey.KeyValue, orig.SubKey.KeyValue) {
		t.Errorf("subkey not preserved: %+v", got.SubKey)
	}
}

// TestETypeInfo2RoundTrip marshals a PA-ETYPE-INFO2 list (etype + salt +
// s2kparams) and confirms each entry survives, mirroring what the KDC returns in
// a PREAUTH_REQUIRED error's e-data.
func TestETypeInfo2RoundTrip(t *testing.T) {
	orig := ETypeInfo2{
		{EType: ETypeAES256CTSHMACSHA196, Salt: "CORP.LOCALalice", S2KParams: []byte{0x00, 0x00, 0x80, 0x00}},
		{EType: ETypeRC4HMAC, Salt: ""},
	}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got ETypeInfo2
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("entry count: %d", len(got))
	}
	if got[0].EType != ETypeAES256CTSHMACSHA196 || got[0].Salt != "CORP.LOCALalice" ||
		!bytes.Equal(got[0].S2KParams, orig[0].S2KParams) {
		t.Errorf("entry0 not preserved: %+v", got[0])
	}
	if got[1].EType != ETypeRC4HMAC || got[1].Salt != "" {
		t.Errorf("entry1 not preserved: %+v", got[1])
	}
}

// TestPAEncTSEncRoundTrip marshals a PA-ENC-TIMESTAMP body and confirms the
// timestamp and microseconds round-trip.
func TestPAEncTSEncRoundTrip(t *testing.T) {
	orig := &PAEncTSEnc{PATimestamp: tstTime, PAUSec: 424242}
	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got PAEncTSEnc
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !got.PATimestamp.Equal(tstTime) || got.PAUSec != 424242 {
		t.Errorf("PA-ENC-TS not preserved: %v/%d", got.PATimestamp, got.PAUSec)
	}
}

// TestEncRepPartRoundTrip marshals EncASRepPart and EncTGSRepPart (APPLICATION 25
// and 26) and confirms the session key, nonce, times and service name round-trip.
func TestEncRepPartRoundTrip(t *testing.T) {
	key := EncryptionKey{KeyType: ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x33}, 32)}

	asPart := &EncASRepPart{
		Key:      key,
		Nonce:    0x2A2A,
		Flags:    NewKerberosFlags(TicketFlagForwardable, TicketFlagInitial),
		AuthTime: tstTime,
		EndTime:  tstTime,
		SRealm:   "CORP.LOCAL",
		SName:    tgtSName(),
	}
	wire, err := asPart.Marshal()
	if err != nil {
		t.Fatalf("EncASRepPart.Marshal: %v", err)
	}
	var gotAS EncASRepPart
	if _, err := gotAS.Unmarshal(wire); err != nil {
		t.Fatalf("EncASRepPart.Unmarshal: %v", err)
	}
	if !bytes.Equal(gotAS.Key.KeyValue, key.KeyValue) || gotAS.Nonce != 0x2A2A ||
		gotAS.SRealm != "CORP.LOCAL" || gotAS.SName.NameString[0] != "krbtgt" {
		t.Errorf("EncASRepPart not preserved: %+v", gotAS)
	}
	if !gotAS.AuthTime.Equal(asPart.AuthTime) {
		t.Errorf("authtime: got %v want %v", gotAS.AuthTime, asPart.AuthTime)
	}

	tgsPart := &EncTGSRepPart{Key: key, Nonce: 99, Flags: NewKerberosFlags(TicketFlagForwardable),
		AuthTime: tstTime, EndTime: tstTime, SRealm: "CORP.LOCAL", SName: tgtSName()}
	wire2, err := tgsPart.Marshal()
	if err != nil {
		t.Fatalf("EncTGSRepPart.Marshal: %v", err)
	}
	var gotTGS EncTGSRepPart
	if _, err := gotTGS.Unmarshal(wire2); err != nil {
		t.Fatalf("EncTGSRepPart.Unmarshal: %v", err)
	}
	if gotTGS.Nonce != 99 || !bytes.Equal(gotTGS.Key.KeyValue, key.KeyValue) {
		t.Errorf("EncTGSRepPart not preserved: %+v", gotTGS)
	}
	// The two enc-part flavours differ only by APPLICATION tag (25 vs 26).
	if wire[0] == wire2[0] {
		t.Errorf("AS-REP and TGS-REP enc-part share an APPLICATION tag byte %#x", wire[0])
	}
}
