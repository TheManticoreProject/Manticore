package mswkst

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// NET_COMPUTER_NAME_ARRAY is the list of computer names returned by
// NetrEnumerateComputerNames ([MS-WKST] 2.2.5.20). ComputerNames is a [unique] pointer to
// a conformant array of EntryCount counted Unicode strings. The [MS-WKST] UNICODE_STRING is
// structurally the [MS-DTYP] RPC_UNICODE_STRING, so we reuse msdtyp.RPC_UNICODE_STRING.
type NET_COMPUTER_NAME_ARRAY struct {
	EntryCount    ndr.DWORD
	ComputerNames []msdtyp.RPC_UNICODE_STRING `ndr:"unique,size_is=EntryCount"`
}
