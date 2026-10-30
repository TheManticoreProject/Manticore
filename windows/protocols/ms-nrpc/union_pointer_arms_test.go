package msnrpc

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// unionPointerArmWire builds the NDR20 wire form of a union whose selected arm is a
// [unique] pointer, as carried by a top-level parameter: the discriminant (tagSize
// octets, padded to 4), the arm's referent id, then the deferred referent body.
func unionPointerArmWire(tag uint32, tagSize int, body []byte) []byte {
	var b []byte
	if tagSize == 2 {
		b = binary.LittleEndian.AppendUint16(b, uint16(tag))
		b = append(b, 0, 0) // pad to the 4-octet alignment of the pointer arm
	} else {
		b = binary.LittleEndian.AppendUint32(b, tag)
	}
	b = binary.LittleEndian.AppendUint32(b, 0x00020000) // arm referent id
	return append(b, body...)
}

// wireUnionParam decodes wire as a union carried by value in a parameter structure (the
// shape every NetrLogon* request/response uses), re-encodes it, and fails unless the
// bytes come back identical. A pointer arm modeled as an inline value consumes the
// referent id as the first body field, so the re-encoded form drops 4 octets.
func wireUnionParam[T any](t *testing.T, name string, wire []byte) {
	t.Helper()
	var p struct{ U T }
	if err := ndr.Unmarshal(wire, &p); err != nil {
		t.Errorf("%s: Unmarshal: %v", name, err)
		return
	}
	raw, err := ndr.Marshal(&p)
	if err != nil {
		t.Errorf("%s: Marshal: %v", name, err)
		return
	}
	if !bytes.Equal(raw, wire) {
		t.Errorf("%s: pointer-arm wire mismatch\n got %x\nwant %x", name, raw, wire)
	}
}

// TestUnionPointerArmsWire verifies that the [MS-NRPC] unions whose IDL arms are P*
// pointer typedefs carry a referent id between the discriminant and the arm body.
func TestUnionPointerArmsWire(t *testing.T) {
	// NETLOGON_LEVEL case 2: PNETLOGON_NETWORK_INFO with empty strings (NULL buffers).
	network := make([]byte, 60)
	binary.LittleEndian.PutUint32(network[8:], 0x11223344) // Identity.ParameterControl
	copy(network[36:44], []byte{1, 2, 3, 4, 5, 6, 7, 8})   // LmChallenge
	wireUnionParam[NETLOGON_LEVEL](t, "NETLOGON_LEVEL/network",
		unionPointerArmWire(uint32(NetlogonNetworkInformation), 2, network))

	// NETLOGON_VALIDATION case 5: PNETLOGON_VALIDATION_GENERIC_INFO2 with a 4-byte blob.
	generic := []byte{
		4, 0, 0, 0, // DataLength
		0x04, 0x00, 0x02, 0x00, // ValidationData referent id
		4, 0, 0, 0, // ValidationData maximum_count
		0xde, 0xad, 0xbe, 0xef,
	}
	wireUnionParam[NETLOGON_VALIDATION](t, "NETLOGON_VALIDATION/generic2",
		unionPointerArmWire(uint32(NetlogonValidationGenericInfo2), 2, generic))

	// NETLOGON_CONTROL_QUERY_INFORMATION case 1: PNETLOGON_INFO_1.
	info1 := []byte{0x80, 0, 0, 0, 0x5e, 0x07, 0, 0}
	wireUnionParam[NETLOGON_CONTROL_QUERY_INFORMATION](t, "NETLOGON_CONTROL_QUERY_INFORMATION/1",
		unionPointerArmWire(1, 4, info1))

	// NETLOGON_DOMAIN_INFORMATION case 2: PNETLOGON_LSA_POLICY_INFO with no policy.
	wireUnionParam[NETLOGON_DOMAIN_INFORMATION](t, "NETLOGON_DOMAIN_INFORMATION/2",
		unionPointerArmWire(2, 4, make([]byte, 8)))

	// NETLOGON_WORKSTATION_INFORMATION case 1: PNETLOGON_WORKSTATION_INFO, all NULL.
	wksta := make([]byte, 80)
	binary.LittleEndian.PutUint32(wksta[64:], 0x3) // WorkstationFlags
	wireUnionParam[NETLOGON_WORKSTATION_INFORMATION](t, "NETLOGON_WORKSTATION_INFORMATION/1",
		unionPointerArmWire(1, 4, wksta))
}
