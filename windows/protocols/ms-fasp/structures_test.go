package msfasp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// deep equality. This is the wire-shape acceptance gate for the MS-FASP (RemoteFW) NDR
// structures in the absence of a live Windows Firewall service (the interface requires
// ncacn_ip_tcp with PKT_PRIVACY + SPNEGO, which cannot be driven from unit tests).
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

// TestV1EnumWidth verifies the [v1_enum]-attributed enums (FW_RULE_STATUS,
// FW_PROFILE_TYPE, FW_RULE_STATUS_CLASS, FW_RULE_CATEGORY) are modeled 4 bytes wide,
// not the default 2 bytes — their documented values (e.g. FW_RULE_STATUS_ALL =
// 0xFFFF0000) do not fit in 16 bits.
func TestV1EnumWidth(t *testing.T) {
	type wrapStatus struct{ V FW_RULE_STATUS }
	raw, err := ndr.Marshal(&wrapStatus{V: FW_RULE_STATUS_ALL})
	if err != nil {
		t.Fatalf("marshal FW_RULE_STATUS: %v", err)
	}
	if len(raw) != 4 {
		t.Errorf("FW_RULE_STATUS marshalled to %d bytes, want 4 (v1_enum)", len(raw))
	}
	type wrapProfile struct{ V FW_PROFILE_TYPE }
	raw, err = ndr.Marshal(&wrapProfile{V: FW_PROFILE_TYPE_CURRENT})
	if err != nil {
		t.Fatalf("marshal FW_PROFILE_TYPE: %v", err)
	}
	if len(raw) != 4 {
		t.Errorf("FW_PROFILE_TYPE marshalled to %d bytes, want 4 (v1_enum)", len(raw))
	}
	// A plain (non-v1_enum) FW_DATA_TYPE stays 2 bytes.
	type wrapData struct{ V FW_DATA_TYPE }
	raw, err = ndr.Marshal(&wrapData{V: FW_DATA_TYPE_UINT32})
	if err != nil {
		t.Fatalf("marshal FW_DATA_TYPE: %v", err)
	}
	if len(raw) != 2 {
		t.Errorf("FW_DATA_TYPE marshalled to %d bytes, want 2 (plain NDR enum)", len(raw))
	}
	roundTrip(t, "FW_RULE_STATUS_CLASS", wrapEnumClass{V: FW_RULE_STATUS_CLASS_ERROR})
}

type wrapEnumClass struct{ V FW_RULE_STATUS_CLASS }

// TestFW_ICMP_TYPE_CODE_LIST exercises a [unique] pointer to a conformant array
// (PEntries, sized by DwNumEntries) — the common enumerated-buffer shape.
func TestFW_ICMP_TYPE_CODE_LIST(t *testing.T) {
	roundTrip(t, "icmp/two", FW_ICMP_TYPE_CODE_LIST{
		DwNumEntries: 2,
		PEntries:     []FW_ICMP_TYPE_CODE{{BType: 8, WCode: 0}, {BType: 3, WCode: 4}},
	})
	roundTrip(t, "icmp/empty", FW_ICMP_TYPE_CODE_LIST{DwNumEntries: 0, PEntries: []FW_ICMP_TYPE_CODE{}})
}

// TestFW_MATCH_VALUE exercises the FW_DATA_TYPE-discriminated union: an integer arm
// (case=3) and the Unicode-string arm (case=5), each selected by its Tag.
func TestFW_MATCH_VALUE(t *testing.T) {
	roundTrip(t, "match/uint32", FW_MATCH_VALUE{
		Type:  FW_DATA_TYPE_UINT32,
		Field: FW_MATCH_VALUE_Field{Tag: FW_DATA_TYPE_UINT32, UInt32: 0xDEADBEEF},
	})
	roundTrip(t, "match/string", FW_MATCH_VALUE{
		Type: FW_DATA_TYPE_UNICODE_STRING,
		Field: FW_MATCH_VALUE_Field{
			Tag: FW_DATA_TYPE_UNICODE_STRING,
			Str: FW_MATCH_VALUE_STRING{WszString: wstr("example.com")},
		},
	})
}

// TestFW_AUTH_INFO exercises the FW_AUTH_METHOD-discriminated union: the Kerberos arm
// (case=2, principal-id strings) and the certificate arm (case=5, a cert pair).
func TestFW_AUTH_INFO(t *testing.T) {
	roundTrip(t, "authinfo/kerb", FW_AUTH_INFO{
		AuthMethod: FW_AUTH_METHOD_MACHINE_KERB,
		Field: FW_AUTH_INFO_Field{
			Tag:         FW_AUTH_METHOD_MACHINE_KERB,
			MachineKerb: FW_AUTH_INFO_KERB{WszMyId: wstr("me@corp"), WszPeerId: wstr("peer@corp")},
		},
		DwAuthInfoFlags: 1,
	})
	roundTrip(t, "authinfo/cert", FW_AUTH_INFO{
		AuthMethod: FW_AUTH_METHOD_MACHINE_CERT,
		Field: FW_AUTH_INFO_Field{
			Tag: FW_AUTH_METHOD_MACHINE_CERT,
			MachineCert: FW_AUTH_INFO_CERT{
				MyCert:   FW_CERT_INFO{SubjectName: FW_BYTE_BLOB{DwSize: 3, Blob: []byte{1, 2, 3}}, DwCertFlags: 0},
				PeerCert: FW_CERT_INFO{SubjectName: FW_BYTE_BLOB{DwSize: 0, Blob: []byte{}}, DwCertFlags: 0},
			},
		},
	})
}

// TestFW_CRYPTO_SET_Field exercises the FW_IPSEC_PHASE-discriminated union: the Phase 1
// arm (case=1) carrying a conformant suite array, and the Phase 2 arm (case=2).
func TestFW_CRYPTO_SET_Field(t *testing.T) {
	roundTrip(t, "crypto/phase1", FW_CRYPTO_SET_Field{
		Tag: FW_IPSEC_PHASE_1,
		Phase1: FW_CRYPTO_SET_PHASE1{
			WFlags:            2,
			DwNumPhase1Suites: 1,
			PPhase1Suites:     []FW_PHASE1_CRYPTO_SUITE{{KeyExchange: 1, Encryption: 2, Hash: 3, DwP1CryptoSuiteFlags: 0}},
			DwTimeOutMinutes:  60,
			DwTimeOutSessions: 100,
		},
	})
	roundTrip(t, "crypto/phase2", FW_CRYPTO_SET_Field{
		Tag: FW_IPSEC_PHASE_2,
		Phase2: FW_CRYPTO_SET_PHASE2{
			Pfs:               FW_PHASE2_CRYPTO_PFS_DH2,
			DwNumPhase2Suites: 1,
			PPhase2Suites:     []FW_PHASE2_CRYPTO_SUITE{{Protocol: 1, EspHash: 2, Encryption: 3}},
		},
	})
}

// TestFW_PROFILE_CONFIG_VALUE exercises the FW_PROFILE_CONFIG-discriminated union: the
// DWORD-pointer arm (case=1), the path-string arm (case=9), and the interface-LUID-list
// arm (case=15, a [unique] pointer to a struct with a conformant GUID array).
func TestFW_PROFILE_CONFIG_VALUE(t *testing.T) {
	dw := ndr.DWORD(1)
	roundTrip(t, "cfg/dword", FW_PROFILE_CONFIG_VALUE{
		Tag:         FW_PROFILE_CONFIG_ENABLE_FW,
		PdwEnableFW: &dw,
	})
	roundTrip(t, "cfg/path", FW_PROFILE_CONFIG_VALUE{
		Tag:    FW_PROFILE_CONFIG_LOG_FILE_PATH,
		WszStr: wstr(`%windir%\fw.log`),
	})
	roundTrip(t, "cfg/interfaces", FW_PROFILE_CONFIG_VALUE{
		Tag: FW_PROFILE_CONFIG_DISABLED_INTERFACES,
		PDisabledInterfaces: &FW_INTERFACE_LUIDS{
			DwNumLUIDs: 1,
			PLUIDs:     []guid.GUID{{A: 0x11111111, B: 0x2222, C: 0x3333, D: 0x4444, E: 0x555566667777}},
		},
	})
}

// TestFW_NETWORK_NAMES exercises a conformant array of [unique] wide-string pointers.
func TestFW_NETWORK_NAMES(t *testing.T) {
	roundTrip(t, "netnames/two", FW_NETWORK_NAMES{
		DwNumEntries: 2,
		WszNames:     []*ndr.WSTR{wstr("Home"), wstr("Work")},
	})
}

// TestFW_RULE exercises the central FW_RULE structure: a self-referential [unique]
// PNext (a two-element linked list), the wIpProtocol port union (TCP arm, case=6), and
// the optional PMetaData conformant array.
func TestFW_RULE(t *testing.T) {
	tail := &FW_RULE{
		WSchemaVersion: 0x0210,
		WszRuleId:      wstr("rule-2"),
		WIpProtocol:    1,
		Field:          FW_RULE_Field{Tag: 1, V4TypeCodeList: FW_ICMP_TYPE_CODE_LIST{DwNumEntries: 0, PEntries: []FW_ICMP_TYPE_CODE{}}},
	}
	roundTrip(t, "rule/list", FW_RULE{
		PNext:          tail,
		WSchemaVersion: 0x0210,
		WszRuleId:      wstr("rule-1"),
		WszName:        wstr("Allow HTTP"),
		WIpProtocol:    6,
		Field: FW_RULE_Field{
			Tag:      6,
			PortsTCP: FW_RULE_PROTOCOL_PORTS{},
		},
		Reserved:  0,
		PMetaData: []FW_OBJECT_METADATA(nil),
	})
}
