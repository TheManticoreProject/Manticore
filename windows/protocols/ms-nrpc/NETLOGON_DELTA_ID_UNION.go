package msnrpc

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NETLOGON_DELTA_ID_UNION ([MS-NRPC] 2.2.1.5.18) is the discriminated union that identifies
// the object a NETLOGON_DELTA_ENUM refers to; the discriminant (NETLOGON_DELTA_TYPE) precedes
// the selected arm ([C706] 14.3.8).
//
// The IDL groups many delta types onto three arms:
//   - Rid  (ULONG)      : cases 1-12, 20, 21 (account/group/alias deltas)
//   - Sid  (PRPC_SID)   : cases 13-17 (LSA policy/trusted-domain/account deltas)
//   - Name ([string]*)  : cases 18, 19 (LSA secret deltas)
//
// The declarative codec matches a single numeric value per `case=` tag, so each label that
// maps to an arm needs its own field of that arm's type; exactly one field is on the wire
// for a given discriminant. Consumers select the populated field by switching on Tag.
type NETLOGON_DELTA_ID_UNION struct {
	Tag NETLOGON_DELTA_TYPE `ndr:"switch"`

	// Rid arm — a bare ULONG object relative identifier.
	RidAddOrChangeDomain     ndr.DWORD `ndr:"case=1"`
	RidAddOrChangeGroup      ndr.DWORD `ndr:"case=2"`
	RidDeleteGroup           ndr.DWORD `ndr:"case=3"`
	RidRenameGroup           ndr.DWORD `ndr:"case=4"`
	RidAddOrChangeUser       ndr.DWORD `ndr:"case=5"`
	RidDeleteUser            ndr.DWORD `ndr:"case=6"`
	RidRenameUser            ndr.DWORD `ndr:"case=7"`
	RidChangeGroupMembership ndr.DWORD `ndr:"case=8"`
	RidAddOrChangeAlias      ndr.DWORD `ndr:"case=9"`
	RidDeleteAlias           ndr.DWORD `ndr:"case=10"`
	RidRenameAlias           ndr.DWORD `ndr:"case=11"`
	RidChangeAliasMembership ndr.DWORD `ndr:"case=12"`
	RidDeleteGroupByName     ndr.DWORD `ndr:"case=20"`
	RidDeleteUserByName      ndr.DWORD `ndr:"case=21"`

	// Sid arm — a [unique] pointer to an RPC_SID (IDL PRPC_SID).
	SidAddOrChangeLsaPolicy  *dtyp.RPC_SID `ndr:"case=13,unique"`
	SidAddOrChangeLsaTDomain *dtyp.RPC_SID `ndr:"case=14,unique"`
	SidDeleteLsaTDomain      *dtyp.RPC_SID `ndr:"case=15,unique"`
	SidAddOrChangeLsaAccount *dtyp.RPC_SID `ndr:"case=16,unique"`
	SidDeleteLsaAccount      *dtyp.RPC_SID `ndr:"case=17,unique"`

	// Name arm — a [unique][string] wide string (IDL wchar_t*).
	NameAddOrChangeLsaSecret *ndr.WSTR `ndr:"case=18,unique"`
	NameDeleteLsaSecret      *ndr.WSTR `ndr:"case=19,unique"`
}
