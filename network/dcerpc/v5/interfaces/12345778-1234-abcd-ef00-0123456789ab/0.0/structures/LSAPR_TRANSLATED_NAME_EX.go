package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_NAME_EX extends LSAPR_TRANSLATED_NAME with a Flags field ([MS-LSAT]
// 2.2.22). Use is an NDR enum (16-bit on the wire); DomainIndex is a signed long.
type LSAPR_TRANSLATED_NAME_EX struct {
	Use         SID_NAME_USE
	Name        dtyp.RPC_UNICODE_STRING
	DomainIndex int32
	Flags       ndr.DWORD
}
