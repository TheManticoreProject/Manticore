package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_TRANSLATED_SID is the translation of a name into a relative id and domain index
// ([MS-LSAT] 2.2.18). Use is an NDR enum (16-bit on the wire); DomainIndex is a signed
// long.
type LSA_TRANSLATED_SID struct {
	Use         SID_NAME_USE
	RelativeId  ndr.DWORD
	DomainIndex int32
}
