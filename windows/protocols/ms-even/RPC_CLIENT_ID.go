package mseven

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// RPC_CLIENT_ID identifies a client process/thread pair ([MS-EVEN] 2.2.6). Both
// members are scalar unsigned longs; there are no pointers or arrays.
type RPC_CLIENT_ID struct {
	UniqueProcess ndr.DWORD
	UniqueThread  ndr.DWORD
}
