package messages

import (
	"encoding/asn1"
	"testing"
	"time"
)

// krbErrorFields decodes a marshaled KRB-ERROR into its context-tagged fields,
// returning for each tag the TLV wrapped by the [tag] EXPLICIT element.
func krbErrorFields(t *testing.T, wire []byte) map[int]asn1.RawValue {
	t.Helper()
	body, _, err := unwrapApplication(wire, MsgTypeError)
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	var elems []asn1.RawValue
	if _, err := asn1.Unmarshal(body, &elems); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	fields := make(map[int]asn1.RawValue)
	for _, e := range elems {
		var inner asn1.RawValue
		if _, err := asn1.Unmarshal(e.Bytes, &inner); err != nil {
			t.Fatalf("decode [%d]: %v", e.Tag, err)
		}
		fields[e.Tag] = inner
	}
	return fields
}

// TestKRBErrorMarshalUsesKerberosStringAndTime checks that KRBError.Marshal
// encodes crealm, realm and the sname components as GeneralString (RFC 4120
// Realm / KerberosString) and stime as a UTC KerberosTime ending in "Z".
func TestKRBErrorMarshalUsesKerberosStringAndTime(t *testing.T) {
	cet := time.FixedZone("CET", 3600)
	in := KRBError{
		STime:     time.Date(2026, 10, 31, 13, 4, 5, 678000000, cet),
		SUSec:     678000,
		ErrorCode: ErrWrongRealm,
		CRealm:    "CHILD.CORP.LOCAL",
		Realm:     "CORP.LOCAL",
		SName:     PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
	}
	wire, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	fields := krbErrorFields(t, wire)

	var sname struct {
		NameType   int             `asn1:"explicit,tag:0"`
		NameString []asn1.RawValue `asn1:"explicit,tag:1"`
	}
	if _, err := asn1.Unmarshal(fields[10].FullBytes, &sname); err != nil || len(sname.NameString) != 2 {
		t.Fatalf("decode sname: %v", err)
	}
	for name, v := range map[string]asn1.RawValue{"crealm": fields[7], "realm": fields[9], "sname[0]": sname.NameString[0], "sname[1]": sname.NameString[1]} {
		if v.Class != asn1.ClassUniversal || v.Tag != asn1.TagGeneralString {
			t.Errorf("%s encoded with tag %d, want GeneralString (%d)", name, v.Tag, asn1.TagGeneralString)
		}
	}
	if stime := fields[4]; stime.Tag != asn1.TagGeneralizedTime || string(stime.Bytes) != "20261031120405Z" {
		t.Errorf("stime = tag %d %q, want GeneralizedTime \"20261031120405Z\"", stime.Tag, stime.Bytes)
	}

	var back KRBError
	if _, err := back.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if back.CRealm != in.CRealm || back.Realm != in.Realm || !back.STime.Equal(in.STime.Truncate(time.Second)) {
		t.Errorf("round trip = %+v", back)
	}

	// An empty crealm stays omitted.
	in.CRealm = ""
	wire, err = in.Marshal()
	if err != nil {
		t.Fatalf("Marshal without crealm: %v", err)
	}
	if _, ok := krbErrorFields(t, wire)[7]; ok {
		t.Error("empty crealm was encoded")
	}
}
