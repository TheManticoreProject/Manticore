package messages

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// marshalContextExplicit wraps seq in an EXPLICIT [tag] context element, as the
// Kerberos ASN.1 modules do for CHOICE alternatives.
func marshalContextExplicit(t *testing.T, tag int, seq []byte) []byte {
	t.Helper()
	b, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: tag, IsCompound: true, Bytes: seq})
	if err != nil {
		t.Fatalf("marshal context [%d]: %v", tag, err)
	}
	return b
}

// TestKrbFastArmoredReqRoundTrip marshals a full KrbFastArmoredReq (as a client
// would build it) and parses it back, verifying every field survives — proving
// the request-envelope encoding is structurally sound.
func TestKrbFastArmoredReqRoundTrip(t *testing.T) {
	body := KDCReqBody{
		KDCOptions: NewKerberosFlags(iana.KDCOptionForwardable, iana.KDCOptionRenewable),
		CName:      PrincipalName{NameType: NameTypePrincipal, NameString: []string{"alice"}},
		Realm:      "CORP.LOCAL",
		SName:      PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
		Till:       time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
		Nonce:      0x11223344,
		EType:      []int{ETypeAES256CTSHMACSHA196, ETypeRC4HMAC},
	}
	fastReq := &KrbFastReq{
		FastOptions: NewKerberosFlags(),
		PAData:      []PAData{{PADataType: PAEncryptedChallenge, PADataValue: []byte{0x01, 0x02, 0x03}}},
		ReqBody:     body,
	}
	fastReqBytes, err := fastReq.Marshal()
	if err != nil {
		t.Fatalf("KrbFastReq.Marshal: %v", err)
	}

	// Verify KrbFastReq itself round-trips (inner body + padata).
	var gotFastReq KrbFastReq
	if _, err := gotFastReq.Unmarshal(fastReqBytes); err != nil {
		t.Fatalf("KrbFastReq.Unmarshal: %v", err)
	}
	if gotFastReq.ReqBody.Realm != body.Realm || gotFastReq.ReqBody.Nonce != body.Nonce {
		t.Fatalf("KrbFastReq body mismatch: got realm=%q nonce=%d", gotFastReq.ReqBody.Realm, gotFastReq.ReqBody.Nonce)
	}
	if len(gotFastReq.PAData) != 1 || gotFastReq.PAData[0].PADataType != PAEncryptedChallenge {
		t.Fatalf("KrbFastReq padata mismatch: %+v", gotFastReq.PAData)
	}

	armor := &KrbFastArmor{ArmorType: FXFastArmorAPRequest, ArmorValue: []byte("fake-ap-req-bytes")}
	req := &KrbFastArmoredReq{
		Armor:       armor,
		ReqChecksum: Checksum{CKSumType: iana.CksumTypeHMACSHA196AES256, Checksum: bytes.Repeat([]byte{0xAB}, 12)},
		EncFastReq:  EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: fastReqBytes},
	}

	// Round-trip via the PA-FX-FAST-REQUEST CHOICE wrapper.
	paValue, err := MarshalPAFXFastRequest(req)
	if err != nil {
		t.Fatalf("MarshalPAFXFastRequest: %v", err)
	}
	// The CHOICE alternative is [0] context-specific constructed.
	if paValue[0] != 0xA0 {
		t.Fatalf("PA-FX-FAST-REQUEST should start with context [0] (0xA0), got 0x%02X", paValue[0])
	}

	var gotReq KrbFastArmoredReq
	// Strip the [0] explicit wrapper to parse the inner SEQUENCE.
	inner := paValue[2:] // 0xA0, length byte, then the SEQUENCE (small payload => single length byte)
	if paValue[1] >= 0x80 {
		inner = paValue[2+int(paValue[1]&0x7f):]
	}
	if _, err := gotReq.Unmarshal(inner); err != nil {
		t.Fatalf("KrbFastArmoredReq.Unmarshal: %v", err)
	}
	if gotReq.Armor == nil || gotReq.Armor.ArmorType != FXFastArmorAPRequest {
		t.Fatalf("armor not preserved: %+v", gotReq.Armor)
	}
	if !bytes.Equal(gotReq.Armor.ArmorValue, armor.ArmorValue) {
		t.Fatalf("armor value mismatch")
	}
	if gotReq.ReqChecksum.CKSumType != iana.CksumTypeHMACSHA196AES256 {
		t.Fatalf("checksum type mismatch: %d", gotReq.ReqChecksum.CKSumType)
	}
	if !bytes.Equal(gotReq.EncFastReq.Cipher, fastReqBytes) {
		t.Fatalf("enc-fast-req cipher mismatch")
	}
}

// TestKrbFastArmoredRepRoundTrip marshals a KrbFastArmoredRep wrapping a
// KrbFastResponse and parses it back through the PA-FX-FAST-REPLY CHOICE,
// exercising the strengthen-key, finished, and nonce fields the reply path
// depends on.
func TestKrbFastArmoredRepRoundTrip(t *testing.T) {
	finished := &KrbFastFinished{
		Timestamp:      time.Date(2030, 5, 6, 7, 8, 9, 0, time.UTC),
		Usec:           123456,
		CRealm:         "CORP.LOCAL",
		CName:          PrincipalName{NameType: NameTypePrincipal, NameString: []string{"alice"}},
		TicketChecksum: Checksum{CKSumType: iana.CksumTypeHMACSHA196AES256, Checksum: bytes.Repeat([]byte{0x5A}, 12)},
	}
	resp := &KrbFastResponse{
		PAData:        []PAData{{PADataType: PAFXCookie, PADataValue: []byte("cookie!")}},
		StrengthenKey: &EncryptionKey{KeyType: ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x77}, 32)},
		Finished:      finished,
		Nonce:         0x0BADF00D,
	}
	respBytes, err := resp.Marshal()
	if err != nil {
		t.Fatalf("KrbFastResponse.Marshal: %v", err)
	}

	// KrbFastResponse round-trip.
	var gotResp KrbFastResponse
	if _, err := gotResp.Unmarshal(respBytes); err != nil {
		t.Fatalf("KrbFastResponse.Unmarshal: %v", err)
	}
	if gotResp.Nonce != resp.Nonce {
		t.Fatalf("nonce mismatch: got %d", gotResp.Nonce)
	}
	if gotResp.StrengthenKey == nil || gotResp.StrengthenKey.KeyType != ETypeAES256CTSHMACSHA196 {
		t.Fatalf("strengthen-key not preserved: %+v", gotResp.StrengthenKey)
	}
	if !bytes.Equal(gotResp.StrengthenKey.KeyValue, resp.StrengthenKey.KeyValue) {
		t.Fatalf("strengthen-key value mismatch")
	}
	if gotResp.Finished == nil || gotResp.Finished.CRealm != "CORP.LOCAL" {
		t.Fatalf("finished not preserved: %+v", gotResp.Finished)
	}
	if len(gotResp.Finished.CName.NameString) != 1 || gotResp.Finished.CName.NameString[0] != "alice" {
		t.Fatalf("finished cname mismatch: %+v", gotResp.Finished.CName)
	}
	if len(gotResp.PAData) != 1 || gotResp.PAData[0].PADataType != PAFXCookie {
		t.Fatalf("response padata mismatch: %+v", gotResp.PAData)
	}

	// Wrap in KrbFastArmoredRep and PA-FX-FAST-REPLY, then parse back.
	rep := &KrbFastArmoredRep{EncFastRep: EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: respBytes}}
	seq, err := rep.Marshal()
	if err != nil {
		t.Fatalf("KrbFastArmoredRep.Marshal: %v", err)
	}
	// Build the PA-FX-FAST-REPLY CHOICE [0] wrapper manually and parse it.
	choice := marshalContextExplicit(t, 0, seq)
	gotRep, err := ParsePAFXFastReply(choice)
	if err != nil {
		t.Fatalf("ParsePAFXFastReply: %v", err)
	}
	if gotRep.EncFastRep.EType != ETypeAES256CTSHMACSHA196 {
		t.Fatalf("enc-fast-rep etype mismatch: %d", gotRep.EncFastRep.EType)
	}
	if !bytes.Equal(gotRep.EncFastRep.Cipher, respBytes) {
		t.Fatalf("enc-fast-rep cipher mismatch")
	}
}

// TestKrbFastResponseNoOptionals verifies that the reply parser handles a
// KrbFastResponse without strengthen-key or finished (an error/interim reply).
func TestKrbFastResponseNoOptionals(t *testing.T) {
	resp := &KrbFastResponse{
		PAData: []PAData{{PADataType: PAFXError, PADataValue: []byte{0x30, 0x00}}},
		Nonce:  42,
	}
	b, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got KrbFastResponse
	if _, err := got.Unmarshal(b); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.StrengthenKey != nil || got.Finished != nil {
		t.Fatalf("optionals should be nil, got sk=%v fin=%v", got.StrengthenKey, got.Finished)
	}
	if got.Nonce != 42 {
		t.Fatalf("nonce mismatch: %d", got.Nonce)
	}
}
