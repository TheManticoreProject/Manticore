package msraiw

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WINSINTF_BIND_DATA_T is the binding information for the WINSIF_HANDLE customized
// ([handle]) binding handle ([MS-RAIW] 2.2.1.5). Its members are the [string] LPSTR
// server address and pipe name, modeled as [unique] ASCII strings. A customized handle
// both selects the binding and is transmitted to the server as a normal [in] data
// parameter ([C706] Customized Handles), so this structure appears on the wire in
// R_WinsGetBrowserNames (opnum 17) and R_WinsStatusWHdl (opnum 20).
type WINSINTF_BIND_DATA_T struct {
	FTcpIp     ndr.DWORD
	PServerAdd *ndr.STR `ndr:"unique"`
	PPipeName  *ndr.STR `ndr:"unique"`
}
