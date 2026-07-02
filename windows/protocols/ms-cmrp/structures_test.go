package mscmrp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in. This is the wire-shape acceptance gate for the
// MS-CMRP (ClusAPI) NDR structures in the absence of a live cluster.
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

// TestCLUSDSK_DISKID_Signature exercises the [switch_is(DiskIdType)] union selecting
// the DiskIdSignature arm (a 4-byte disk signature). The inline discriminant is a
// 16-bit CLUSDSK_DISKID_ENUM and must match the sibling DiskIdType.
func TestCLUSDSK_DISKID_Signature(t *testing.T) {
	roundTrip(t, "CLUSDSK_DISKID/signature", CLUSDSK_DISKID{
		DiskIdType: DiskIdSignature,
		Field: CLUSDSK_DISKID_Field{
			Tag:           DiskIdSignature,
			DiskSignature: 0x12345678,
		},
	})
}

// TestCLUSDSK_DISKID_Guid exercises the DiskIdGuid arm (a 16-byte GUID) of the same
// union — the second arm with a larger alignment.
func TestCLUSDSK_DISKID_Guid(t *testing.T) {
	roundTrip(t, "CLUSDSK_DISKID/guid", CLUSDSK_DISKID{
		DiskIdType: DiskIdGuid,
		Field: CLUSDSK_DISKID_Field{
			Tag:      DiskIdGuid,
			DiskGuid: guid.GUID{A: 0xc681d488, B: 0xd850, C: 0x11d0, D: 0x8c52, E: 0x00c04fd90f7e},
		},
	})
}

// TestENUM_LIST exercises a [unique] pointer to a conformant array (Entry, sized by
// EntryCount) of ENUM_ENTRY structures that each carry a [unique] string pointer —
// the shape returned by the ApiCreateEnum family.
func TestENUM_LIST(t *testing.T) {
	roundTrip(t, "ENUM_LIST/two", ENUM_LIST{
		EntryCount: 2,
		Entry: []ENUM_ENTRY{
			{Type: 1, Name: wstr("Cluster Group")},
			{Type: 2, Name: wstr("Available Storage")},
		},
	})
	// An embedded conformant array is always present (no null referent): an empty
	// list is EntryCount 0 with a zero-length slice, which is what it decodes back to.
	roundTrip(t, "ENUM_LIST/empty", ENUM_LIST{EntryCount: 0, Entry: []ENUM_ENTRY{}})
}

// TestGROUP_ENUM_LIST exercises the group enumeration list: an array of entries that
// each hold several [unique] string pointers plus two counted property blobs
// ([unique] byte arrays sized by their preceding count fields).
func TestGROUP_ENUM_LIST(t *testing.T) {
	roundTrip(t, "GROUP_ENUM_LIST", GROUP_ENUM_LIST{
		EntryCount: 1,
		Entry: []GROUP_ENUM_ENTRY{{
			Name:           wstr("Cluster Group"),
			Id:             wstr("1f8d9c3a"),
			DwState:        2,
			Owner:          wstr("NODE1"),
			DwFlags:        0,
			CbProperties:   4,
			Properties:     []uint8{0xde, 0xad, 0xbe, 0xef},
			CbRoProperties: 2,
			RoProperties:   []uint8{0x01, 0x02},
		}},
	})
}

// TestRPC_SECURITY_DESCRIPTOR exercises the counted, [unique]/varying byte blob
// (size_is(cbIn), length_is(cbOut)) and its containing RPC_SECURITY_ATTRIBUTES.
func TestRPC_SECURITY_DESCRIPTOR(t *testing.T) {
	roundTrip(t, "RPC_SECURITY_DESCRIPTOR", RPC_SECURITY_DESCRIPTOR{
		LpSecurityDescriptor:    []uint8{0x01, 0x00, 0x04, 0x80},
		CbInSecurityDescriptor:  4,
		CbOutSecurityDescriptor: 4,
	})
	roundTrip(t, "RPC_SECURITY_ATTRIBUTES", RPC_SECURITY_ATTRIBUTES{
		NLength: 12,
		RpcSecurityDescriptor: RPC_SECURITY_DESCRIPTOR{
			LpSecurityDescriptor:    []uint8{0xaa, 0xbb},
			CbInSecurityDescriptor:  2,
			CbOutSecurityDescriptor: 2,
		},
		BInheritHandle: 1,
	})
}

// TestNOTIFICATION_DATA_RPC exercises the notification payload: a nested struct with a
// 64-bit FilterFlags, a [unique] counted buffer, and four [unique] string pointers.
func TestNOTIFICATION_DATA_RPC(t *testing.T) {
	roundTrip(t, "NOTIFICATION_DATA_RPC", NOTIFICATION_DATA_RPC{
		FilterAndType: NOTIFY_FILTER_AND_TYPE_RPC{DwObjectType: 3, FilterFlags: 0x0000000100000002},
		Buffer:        []uint8{0x11, 0x22, 0x33},
		DwBufferSize:  3,
		ObjectId:      wstr("res-guid"),
		ParentId:      wstr("grp-guid"),
		Name:          wstr("Physical Disk"),
		Type:          wstr("Storage"),
	})
}

// TestIDL_CLUSTER_SET_PASSWORD_STATUS exercises a plain struct with a 1-byte BOOLEAN
// field alongside DWORDs.
func TestIDL_CLUSTER_SET_PASSWORD_STATUS(t *testing.T) {
	roundTrip(t, "IDL_CLUSTER_SET_PASSWORD_STATUS", IDL_CLUSTER_SET_PASSWORD_STATUS{
		NodeId:       7,
		SetAttempted: true,
		ReturnStatus: 0,
	})
}

// TestCLUSTER_OPERATIONAL_VERSION_INFO exercises the fixed all-scalar version struct.
func TestCLUSTER_OPERATIONAL_VERSION_INFO(t *testing.T) {
	roundTrip(t, "CLUSTER_OPERATIONAL_VERSION_INFO", CLUSTER_OPERATIONAL_VERSION_INFO{
		DwSize:                  20,
		DwClusterHighestVersion: 0x000600A5,
		DwClusterLowestVersion:  0x000600A5,
		DwFlags:                 0,
		DwReserved:              0,
	})
}
