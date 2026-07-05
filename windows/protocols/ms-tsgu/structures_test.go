package mstsgu

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTripTSGU marshals in, unmarshals into a fresh value of the same type, and
// asserts the result is deeply equal to in.
func roundTripTSGU[T any](t *testing.T, name string, in T) {
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

func wstr(s string) *ndr.WSTR {
	w := ndr.WSTR(s)
	return &w
}

// TestTSG_PACKET_HEADER_RoundTrip exercises a plain two-field scalar struct.
func TestTSG_PACKET_HEADER_RoundTrip(t *testing.T) {
	roundTripTSGU(t, "TSG_PACKET_HEADER", TSG_PACKET_HEADER{ComponentId: 0x4844, PacketId: 0x5643})
}

// TestTSG_PACKET_CAPABILITIES_RoundTrip exercises a struct embedding a union whose
// selected arm is a value (TSG_CAPABILITY_TYPE_NAP). The union's inline discriminant
// duplicates the parent CapabilityType field.
func TestTSG_PACKET_CAPABILITIES_RoundTrip(t *testing.T) {
	in := TSG_PACKET_CAPABILITIES{
		CapabilityType: 0x00000001, // TSG_CAPABILITY_TYPE_NAP
		TSGPacket: TSG_CAPABILITIES_UNION{
			Tag:       0x00000001,
			TSGCapNap: TSG_CAPABILITY_NAP{Capabilities: 0x3},
		},
	}
	roundTripTSGU(t, "TSG_PACKET_CAPABILITIES", in)
}

// TestTSG_PACKET_VERSIONCAPS_RoundTrip exercises a [unique] pointer to a conformant
// array of capability structs.
func TestTSG_PACKET_VERSIONCAPS_RoundTrip(t *testing.T) {
	in := TSG_PACKET_VERSIONCAPS{
		TsgHeader: TSG_PACKET_HEADER{ComponentId: 0x5644, PacketId: 0x5643},
		TSGCaps: []TSG_PACKET_CAPABILITIES{
			{
				CapabilityType: 0x00000001,
				TSGPacket:      TSG_CAPABILITIES_UNION{Tag: 0x00000001, TSGCapNap: TSG_CAPABILITY_NAP{Capabilities: 0x1}},
			},
		},
		NumCapabilities:        1,
		MajorVersion:           1,
		MinorVersion:           1,
		QuarantineCapabilities: 0,
	}
	roundTripTSGU(t, "TSG_PACKET_VERSIONCAPS", in)
}

// TestTSG_PACKET_QUARREQUEST_RoundTrip exercises a [string] conformant-varying wide
// array (MachineName) alongside a [unique] conformant byte blob (Data).
func TestTSG_PACKET_QUARREQUEST_RoundTrip(t *testing.T) {
	in := TSG_PACKET_QUARREQUEST{
		Flags:       0,
		MachineName: []uint16{'H', 'O', 'S', 'T', 0},
		NameLength:  5,
		Data:        []uint8{0xde, 0xad, 0xbe, 0xef},
		DataLen:     4,
	}
	roundTripTSGU(t, "TSG_PACKET_QUARREQUEST", in)
}

// TestTSG_PACKET_RESPONSE_RoundTrip exercises a conformant byte blob followed by an
// embedded fixed struct of BOOLs.
func TestTSG_PACKET_RESPONSE_RoundTrip(t *testing.T) {
	in := TSG_PACKET_RESPONSE{
		Flags:           1,
		Reserved:        0,
		ResponseData:    []uint8{1, 2, 3},
		ResponseDataLen: 3,
		RedirectionFlags: TSG_REDIRECTION_FLAGS{
			EnableAllRedirections:        1,
			DisableAllRedirections:       0,
			DriveRedirectionDisabled:     1,
			PrinterRedirectionDisabled:   0,
			PortRedirectionDisabled:      0,
			Reserved:                     0,
			ClipboardRedirectionDisabled: 0,
			PnpRedirectionDisabled:       0,
		},
	}
	roundTripTSGU(t, "TSG_PACKET_RESPONSE", in)
}

// TestTSG_PACKET_QUARENC_RESPONSE_RoundTrip exercises a [string] wide array, an
// embedded GUID, and a [unique] pointer to a version-caps struct.
func TestTSG_PACKET_QUARENC_RESPONSE_RoundTrip(t *testing.T) {
	in := TSG_PACKET_QUARENC_RESPONSE{
		Flags:         0,
		CertChainLen:  4,
		CertChainData: []uint16{'C', 'E', 'R', 0},
		Nonce:         dtyp.GUID{Data1: 0x11223344, Data2: 0x5566, Data3: 0x7788, Data4: [8]byte{0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00}},
		VersionCaps: &TSG_PACKET_VERSIONCAPS{
			TsgHeader:       TSG_PACKET_HEADER{ComponentId: 0x5644, PacketId: 0x5643},
			TSGCaps:         nil,
			NumCapabilities: 0,
			MajorVersion:    1,
			MinorVersion:    1,
		},
	}
	roundTripTSGU(t, "TSG_PACKET_QUARENC_RESPONSE", in)
}

// TestTSG_PACKET_VersionCapsArm_RoundTrip exercises the top-level TSG_PACKET whose
// union selects the [unique] pointer VERSIONCAPS arm.
func TestTSG_PACKET_VersionCapsArm_RoundTrip(t *testing.T) {
	in := TSG_PACKET{
		PacketId: 0x00005643, // TSG_PACKET_TYPE_VERSIONCAPS
		TSGPacket: TSG_PACKET_TYPE_UNION{
			Tag: 0x00005643,
			PacketVersionCaps: &TSG_PACKET_VERSIONCAPS{
				TsgHeader:       TSG_PACKET_HEADER{ComponentId: 0x5644, PacketId: 0x5643},
				NumCapabilities: 0,
				MajorVersion:    1,
				MinorVersion:    1,
			},
		},
	}
	roundTripTSGU(t, "TSG_PACKET/VERSIONCAPS", in)
}

// TestTSG_PACKET_QuarRequestArm_RoundTrip exercises the TSG_PACKET union selecting the
// QUARREQUEST arm (a pointer to a struct carrying its own varying string).
func TestTSG_PACKET_QuarRequestArm_RoundTrip(t *testing.T) {
	in := TSG_PACKET{
		PacketId: 0x00005152, // TSG_PACKET_TYPE_QUARREQUEST
		TSGPacket: TSG_PACKET_TYPE_UNION{
			Tag: 0x00005152,
			PacketQuarRequest: &TSG_PACKET_QUARREQUEST{
				Flags:       0,
				MachineName: []uint16{'M', 0},
				NameLength:  2,
				Data:        nil,
				DataLen:     0,
			},
		},
	}
	roundTripTSGU(t, "TSG_PACKET/QUARREQUEST", in)
}

// TestTSG_PACKET_MSG_RESPONSE_RoundTrip exercises the message-response struct whose
// embedded union selects the [unique] pointer consent-message arm.
func TestTSG_PACKET_MSG_RESPONSE_RoundTrip(t *testing.T) {
	in := TSG_PACKET_MSG_RESPONSE{
		MsgID:        7,
		MsgType:      0x00000001, // TSG_ASYNC_MESSAGE_CONSENT_MESSAGE
		IsMsgPresent: 1,
		MessagePacket: TSG_PACKET_TYPE_MESSAGE_UNION{
			Tag: 0x00000001,
			ConsentMessage: &TSG_PACKET_STRING_MESSAGE{
				IsDisplayMandatory: 1,
				IsConsentMandatory: 1,
				MsgBytes:           3,
				MsgBuffer:          []uint16{'H', 'i', 0},
			},
		},
	}
	roundTripTSGU(t, "TSG_PACKET_MSG_RESPONSE", in)
}

// TestTSG_PACKET_REAUTH_RoundTrip exercises the reauth struct: a 64-bit tunnel context,
// a discriminant, and an embedded union selecting a [unique] pointer AUTH arm.
func TestTSG_PACKET_REAUTH_RoundTrip(t *testing.T) {
	in := TSG_PACKET_REAUTH{
		TunnelContext: 0x0123456789abcdef,
		PacketId:      0x00004054, // TSG_PACKET_TYPE_AUTH
		TSGInitialPacket: TSG_INITIAL_PACKET_TYPE_UNION{
			Tag: 0x00004054,
			PacketAuth: &TSG_PACKET_AUTH{
				TSGVersionCaps: TSG_PACKET_VERSIONCAPS{
					TsgHeader:    TSG_PACKET_HEADER{ComponentId: 0x5644, PacketId: 0x5643},
					MajorVersion: 1,
					MinorVersion: 1,
				},
				CookieLen: 2,
				Cookie:    []uint8{0xaa, 0xbb},
			},
		},
	}
	roundTripTSGU(t, "TSG_PACKET_REAUTH", in)
}

// TestTSENDPOINTINFO_RoundTrip exercises a [unique] pointer to a conformant array of
// [unique] wide-string pointers, plus a shorter alternate-names array.
func TestTSENDPOINTINFO_RoundTrip(t *testing.T) {
	in := TSENDPOINTINFO{
		ResourceName:              []*ndr.WSTR{wstr("server1.contoso.com")},
		NumResourceNames:          1,
		AlternateResourceNames:    []*ndr.WSTR{wstr("alt1")},
		NumAlternateResourceNames: 1,
		Port:                      3389,
	}
	roundTripTSGU(t, "TSENDPOINTINFO", in)
}
