package mslsat

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_TRANSLATED_SID_EX2 carries the full SID (not just a relative id) of a
// translation result ([MS-LSAT] 2.2.26). Use is an NDR enum (16-bit on the wire); Sid is
// a [unique] pointer to an RPC_SID; DomainIndex is a signed long.
type LSAPR_TRANSLATED_SID_EX2 struct {
	Use         SID_NAME_USE    `ndr:"enum"`
	Sid         *msdtyp.RPC_SID `ndr:"unique"`
	DomainIndex int32
	Flags       ndr.DWORD
}
