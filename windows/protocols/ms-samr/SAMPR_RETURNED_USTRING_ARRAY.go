package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_RETURNED_USTRING_ARRAY is a counted array of RPC_UNICODE_STRING values
// ([MS-SAMR] 2.2.3.9). Element is a [unique] pointer to a conformant array sized
// by Count.
type SAMPR_RETURNED_USTRING_ARRAY struct {
	Count   ndr.DWORD
	Element []msdtyp.RPC_UNICODE_STRING `ndr:"unique,size_is=Count"`
}
