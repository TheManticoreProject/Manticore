package msdssp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-DSSP NDR
// structures in the absence of a live Directory Services Setup server.
func roundTrip[T any](t *testing.T, name string, in T) []byte {
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
	return raw
}

// sampleGUID exercises every octet slot so a byte-order or truncation bug in the 16-octet
// dtyp.GUID field shows up in the round trip.
var sampleGUID = dtyp.GUID{
	Data1: 0x11223344,
	Data2: 0x5566,
	Data3: 0x7788,
	Data4: [8]byte{0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00},
}

// TestDsroleUpgradeStatusInfo round-trips the upgrade-status arm (a DWORD + a 16-bit enum).
func TestDsroleUpgradeStatusInfo(t *testing.T) {
	roundTrip(t, "DSROLE_UPGRADE_STATUS_INFO", DSROLE_UPGRADE_STATUS_INFO{
		OperationState:      0x0000ABCD,
		PreviousServerState: DsRoleServerPrimary,
	})
}

// TestDsroleOperationStateInfo round-trips the operation-state arm (a single 16-bit enum).
func TestDsroleOperationStateInfo(t *testing.T) {
	roundTrip(t, "DSROLE_OPERATION_STATE_INFO", DSROLE_OPERATION_STATE_INFO{
		OperationState: DsRoleOperationNeedReboot,
	})
}

// TestDsrolerPrimaryDomainInfoBasicFull round-trips the basic arm with all three [unique]
// string pointers present. The GUID must contribute exactly 16 octets on the wire.
func TestDsrolerPrimaryDomainInfoBasicFull(t *testing.T) {
	flat := ndr.WSTR("CONTOSO")
	dns := ndr.WSTR("contoso.com")
	forest := ndr.WSTR("contoso.com")
	roundTrip(t, "DSROLER_PRIMARY_DOMAIN_INFO_BASIC(full)", DSROLER_PRIMARY_DOMAIN_INFO_BASIC{
		MachineRole:      DsRole_RolePrimaryDomainController,
		Flags:            0x01234567,
		DomainNameFlat:   &flat,
		DomainNameDns:    &dns,
		DomainForestName: &forest,
		DomainGuid:       sampleGUID,
	})
}

// TestDsrolerPrimaryDomainInfoBasicNull round-trips the basic arm with all name pointers
// null — the standalone-workstation case, where no domain names are set.
func TestDsrolerPrimaryDomainInfoBasicNull(t *testing.T) {
	roundTrip(t, "DSROLER_PRIMARY_DOMAIN_INFO_BASIC(null)", DSROLER_PRIMARY_DOMAIN_INFO_BASIC{
		MachineRole: DsRole_RoleStandaloneWorkstation,
		Flags:       0,
		DomainGuid:  dtyp.GUID{},
	})
}

// TestDsrolerPrimaryDomainInformationBasic round-trips the union selecting the basic arm.
func TestDsrolerPrimaryDomainInformationBasic(t *testing.T) {
	flat := ndr.WSTR("CONTOSO")
	roundTrip(t, "DSROLER_PRIMARY_DOMAIN_INFORMATION(basic)", DSROLER_PRIMARY_DOMAIN_INFORMATION{
		Tag: DsRolePrimaryDomainInfoBasic,
		DomainInfoBasic: DSROLER_PRIMARY_DOMAIN_INFO_BASIC{
			MachineRole:    DsRole_RoleMemberServer,
			Flags:          0xDEADBEEF,
			DomainNameFlat: &flat,
			DomainGuid:     sampleGUID,
		},
	})
}

// TestDsrolerPrimaryDomainInformationUpgrade round-trips the union selecting the
// upgrade-status arm.
func TestDsrolerPrimaryDomainInformationUpgrade(t *testing.T) {
	roundTrip(t, "DSROLER_PRIMARY_DOMAIN_INFORMATION(upgrade)", DSROLER_PRIMARY_DOMAIN_INFORMATION{
		Tag: DsRoleUpgradeStatus,
		UpgradStatusInfo: DSROLE_UPGRADE_STATUS_INFO{
			OperationState:      1,
			PreviousServerState: DsRoleServerBackup,
		},
	})
}

// TestDsrolerPrimaryDomainInformationOperation round-trips the union selecting the
// operation-state arm.
func TestDsrolerPrimaryDomainInformationOperation(t *testing.T) {
	roundTrip(t, "DSROLER_PRIMARY_DOMAIN_INFORMATION(operation)", DSROLER_PRIMARY_DOMAIN_INFORMATION{
		Tag: DsRoleOperationState,
		OperationStateInfo: DSROLE_OPERATION_STATE_INFO{
			OperationState: DsRoleOperationActive,
		},
	})
}
