package messages

import (
	"bytes"
	"testing"
	"time"
)

// collectGeneralizedTimes walks a DER blob and returns the content octets of
// every GeneralizedTime (universal tag 24, 0x18) primitive it contains. It
// recurses into constructed values so explicitly-tagged timestamps are found.
func collectGeneralizedTimes(der []byte) [][]byte {
	var out [][]byte
	i := 0
	for i < len(der) {
		tag := der[i]
		i++
		if i >= len(der) {
			break
		}
		// Decode the (possibly long-form) length.
		l := int(der[i])
		i++
		if l&0x80 != 0 {
			n := l & 0x7f
			l = 0
			for k := 0; k < n && i < len(der); k++ {
				l = l<<8 | int(der[i])
				i++
			}
		}
		if i+l > len(der) {
			break
		}
		content := der[i : i+l]
		if tag&0x20 != 0 { // constructed: recurse into the content
			out = append(out, collectGeneralizedTimes(content)...)
		} else if tag&0x1f == 0x18 && tag&0xc0 == 0x00 { // universal primitive GeneralizedTime
			out = append(out, content)
		}
		i += l
	}
	return out
}

// assertUTCTimes fails unless every GeneralizedTime in der ends with 'Z' and
// carries no numeric zone offset ('+'/'-'), i.e. the DER form RFC 4120 mandates.
func assertUTCTimes(t *testing.T, der []byte, want int) {
	t.Helper()
	times := collectGeneralizedTimes(der)
	if len(times) != want {
		t.Fatalf("expected %d GeneralizedTime value(s), found %d: %q", want, len(times), times)
	}
	for _, tv := range times {
		if len(tv) == 0 || tv[len(tv)-1] != 'Z' {
			t.Errorf("GeneralizedTime %q does not end with UTC 'Z'", tv)
		}
		if bytes.ContainsAny(tv, "+-") {
			t.Errorf("GeneralizedTime %q carries a numeric zone offset", tv)
		}
	}
}

// TestAuthenticatorMarshalNormalizesCTime confirms a non-UTC CTime is emitted as
// a UTC GeneralizedTime (trailing 'Z', no numeric offset).
func TestAuthenticatorMarshalNormalizesCTime(t *testing.T) {
	loc := time.FixedZone("X", -5*3600)
	a := &Authenticator{
		AVno:   KerberosV5,
		CRealm: "CORP.LOCAL",
		CName:  cliName(),
		CUSec:  42,
		CTime:  time.Date(2026, 7, 10, 12, 0, 0, 0, loc),
	}
	wire, err := a.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	assertUTCTimes(t, wire, 1)
}

// TestKDCReqBodyMarshalNormalizesTill confirms a non-UTC Till is emitted in UTC
// and that an omitted optional From/RTime stays absent from the wire form.
func TestKDCReqBodyMarshalNormalizesTill(t *testing.T) {
	loc := time.FixedZone("X", -5*3600)
	body := KDCReqBody{
		KDCOptions: NewKerberosFlags(1, 3, 8),
		CName:      cliName(),
		Realm:      "CORP.LOCAL",
		SName:      tgtSName(),
		Till:       time.Date(2026, 7, 10, 12, 0, 0, 0, loc),
		Nonce:      0x2A2A2A2A,
		EType:      []int{ETypeAES256CTSHMACSHA196},
	}
	wire, err := EncodeKDCReqBody(body)
	if err != nil {
		t.Fatalf("EncodeKDCReqBody: %v", err)
	}
	// From and RTime are left zero, so only Till must appear.
	assertUTCTimes(t, wire, 1)
}

// TestKDCReqBodyMarshalNormalizesAllTimes confirms that when From, Till and RTime
// are all set in a non-UTC location, all three are emitted in UTC form.
func TestKDCReqBodyMarshalNormalizesAllTimes(t *testing.T) {
	loc := time.FixedZone("X", -5*3600)
	tv := time.Date(2026, 7, 10, 12, 0, 0, 0, loc)
	body := KDCReqBody{
		KDCOptions: NewKerberosFlags(1, 3, 8),
		CName:      cliName(),
		Realm:      "CORP.LOCAL",
		SName:      tgtSName(),
		From:       tv,
		Till:       tv,
		RTime:      tv,
		Nonce:      0x2A2A2A2A,
		EType:      []int{ETypeAES256CTSHMACSHA196},
	}
	wire, err := EncodeKDCReqBody(body)
	if err != nil {
		t.Fatalf("EncodeKDCReqBody: %v", err)
	}
	assertUTCTimes(t, wire, 3)
}

// TestPAEncTSEncMarshalNormalizesTimestamp confirms a non-UTC PATimestamp is
// emitted in UTC form and that Marshal does not mutate the receiver's field.
func TestPAEncTSEncMarshalNormalizesTimestamp(t *testing.T) {
	loc := time.FixedZone("X", -5*3600)
	orig := time.Date(2026, 7, 10, 12, 0, 0, 0, loc)
	p := &PAEncTSEnc{PATimestamp: orig, PAUSec: 42}
	wire, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	assertUTCTimes(t, wire, 1)
	if !p.PATimestamp.Equal(orig) || p.PATimestamp.Location() != loc {
		t.Errorf("Marshal mutated receiver PATimestamp: %v", p.PATimestamp)
	}
}
