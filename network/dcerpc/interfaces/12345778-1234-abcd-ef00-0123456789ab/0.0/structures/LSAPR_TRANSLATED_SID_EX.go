package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRANSLATED_SID_EX extends LSA_TRANSLATED_SID with a Flags field ([MS-LSAT]
// 2.2.24). Use is an NDR enum (16-bit on the wire); DomainIndex is a signed long.
type LSAPR_TRANSLATED_SID_EX struct {
	Use         SID_NAME_USE `ndr:"enum"`
	RelativeId  ndr.DWORD
	DomainIndex int32
	Flags       ndr.DWORD
}
