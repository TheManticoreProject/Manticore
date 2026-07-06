package msnrpc

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// dtypSID builds a representative RPC_SID for union round-trip tests.
func dtypSID() *msdtyp.RPC_SID {
	s, err := msdtyp.ParseSID("S-1-5-21-1004336348-1177238915-682003330-512")
	if err != nil {
		panic(err)
	}
	return &s
}

// roundTrip marshals v, unmarshals the bytes back into a fresh value of the same type, and
// fails unless the result is deeply equal to the input. It is the wire-shape regression
// harness for the protocol's NDR structures.
func roundTrip[T any](t *testing.T, name string, v T) {
	t.Helper()
	raw, err := ndr.Marshal(&v)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(v, out) {
		t.Fatalf("%s: round-trip mismatch\n in: %#v\nout: %#v", name, v, out)
	}
}

// TestNetlogonCredentialWire verifies the 8-octet fixed-array credential marshals verbatim
// with no NDR framing when carried as a struct field (its only use on the wire).
func TestNetlogonCredentialWire(t *testing.T) {
	type wrap struct{ C NETLOGON_CREDENTIAL }
	c := NETLOGON_CREDENTIAL{1, 2, 3, 4, 5, 6, 7, 8}
	raw, err := ndr.Marshal(&wrap{C: c})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(raw, c[:]) {
		t.Fatalf("credential wire = %x, want %x", raw, c[:])
	}
}

// TestNetlogonAuthenticatorWire verifies the authenticator is a 12-octet structure (8-byte
// credential + 4-byte timestamp) with the timestamp little-endian at offset 8.
func TestNetlogonAuthenticatorWire(t *testing.T) {
	a := NETLOGON_AUTHENTICATOR{Credential: NETLOGON_CREDENTIAL{1, 1, 1, 1, 1, 1, 1, 1}, Timestamp: 0x04030201}
	raw, err := ndr.Marshal(&a)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{1, 1, 1, 1, 1, 1, 1, 1, 0x01, 0x02, 0x03, 0x04}
	if !bytes.Equal(raw, want) {
		t.Fatalf("authenticator wire = %x, want %x", raw, want)
	}
}

// TestTrustPasswordWire verifies an all-zero NL_TRUST_PASSWORD marshals to 516 zero octets
// (512-byte buffer + 4-byte length), the empty-password form.
func TestTrustPasswordWire(t *testing.T) {
	raw, err := ndr.Marshal(&NL_TRUST_PASSWORD{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 516 {
		t.Fatalf("trust password length = %d, want 516", len(raw))
	}
	for i, b := range raw {
		if b != 0 {
			t.Fatalf("octet %d = 0x%02x, want 0", i, b)
		}
	}
}

// TestScalarStructRoundTrips exercises the codec on the simple fixed-array and scalar
// structures the protocol builds larger types out of.
func TestScalarStructRoundTrips(t *testing.T) {
	roundTrip(t, "GROUP_MEMBERSHIP", GROUP_MEMBERSHIP{RelativeId: 513, Attributes: 7})
	roundTrip(t, "CYPHER_BLOCK", CYPHER_BLOCK{Data: [8]int8{1, 2, 3, 4, 5, 6, 7, 8}})
	roundTrip(t, "USER_SESSION_KEY", USER_SESSION_KEY{Data: [2]CYPHER_BLOCK{
		{Data: [8]int8{1, 1, 1, 1, 1, 1, 1, 1}},
		{Data: [8]int8{2, 2, 2, 2, 2, 2, 2, 2}},
	}})
	roundTrip(t, "OLD_LARGE_INTEGER", OLD_LARGE_INTEGER{LowPart: 0xdeadbeef, HighPart: -1})
	roundTrip(t, "NLPR_QUOTA_LIMITS", NLPR_QUOTA_LIMITS{
		PagedPoolLimit: 1, NonPagedPoolLimit: 2, MinimumWorkingSetSize: 3,
		MaximumWorkingSetSize: 4, PagefileLimit: 5,
		Reserved: OLD_LARGE_INTEGER{LowPart: 6, HighPart: 7},
	})
}

// TestDomainControllerInfoRoundTrip exercises a struct built entirely of [unique] string
// pointers plus an inline GUID — the DsrGetDcName response shape.
func TestDomainControllerInfoRoundTrip(t *testing.T) {
	dcName := ndr.WSTR(`\\DC01`)
	addr := ndr.WSTR(`\\10.0.0.1`)
	dom := ndr.WSTR("CONTOSO")
	roundTrip(t, "DOMAIN_CONTROLLER_INFOW", DOMAIN_CONTROLLER_INFOW{
		DomainControllerName:        &dcName,
		DomainControllerAddress:     &addr,
		DomainControllerAddressType: 1,
		DomainName:                  &dom,
		Flags:                       0xE00013FD,
	})
}

// TestNetlogonValidationUnion selects the generic-info arm (a [unique] pointer to a
// conformant byte array) and round-trips the discriminated union, exercising the switch
// discriminant plus the unique-pointer-to-conformant-array buffer shape together.
func TestNetlogonValidationUnion(t *testing.T) {
	roundTrip(t, "NETLOGON_VALIDATION/generic2", NETLOGON_VALIDATION{
		Tag: NetlogonValidationGenericInfo2,
		ValidationGeneric2: NETLOGON_VALIDATION_GENERIC_INFO2{
			DataLength:     4,
			ValidationData: []uint8{0xde, 0xad, 0xbe, 0xef},
		},
	})
}

// TestDeltaIdUnion exercises the multi-label NETLOGON_DELTA_ID_UNION: a Rid arm reached by
// a non-first discriminant (case 6, DeleteUser), the Sid arm (a [unique] PRPC_SID pointer),
// and the Name arm (a [unique][string] pointer). Each must select exactly its own field and
// round-trip; a discriminant that mapped to no arm would drop the arm bytes and desync.
func TestDeltaIdUnion(t *testing.T) {
	roundTrip(t, "NETLOGON_DELTA_ID_UNION/rid", NETLOGON_DELTA_ID_UNION{
		Tag: DeleteUser, RidDeleteUser: 1105,
	})
	sid := dtypSID()
	roundTrip(t, "NETLOGON_DELTA_ID_UNION/sid", NETLOGON_DELTA_ID_UNION{
		Tag: AddOrChangeLsaPolicy, SidAddOrChangeLsaPolicy: sid,
	})
	name := ndr.WSTR("$MACHINE.ACC")
	roundTrip(t, "NETLOGON_DELTA_ID_UNION/name", NETLOGON_DELTA_ID_UNION{
		Tag: AddOrChangeLsaSecret, NameAddOrChangeLsaSecret: &name,
	})
}

// TestControlDataInformationUnion exercises the switch_type(DWORD) control union: a
// trusted-domain-name arm reached by FunctionCode 9 (one of the 5/6/9/10 group) and the
// DebugFlag scalar arm.
func TestControlDataInformationUnion(t *testing.T) {
	dom := ndr.WSTR("CONTOSO")
	roundTrip(t, "NETLOGON_CONTROL_DATA_INFORMATION/domain9", NETLOGON_CONTROL_DATA_INFORMATION{
		Tag: 9, TrustedDomainName9: &dom,
	})
	roundTrip(t, "NETLOGON_CONTROL_DATA_INFORMATION/debug", NETLOGON_CONTROL_DATA_INFORMATION{
		Tag: 65534, DebugFlag: 0x1234,
	})
}
