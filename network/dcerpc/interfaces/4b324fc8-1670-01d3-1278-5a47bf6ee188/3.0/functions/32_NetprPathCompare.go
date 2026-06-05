package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netprPathCompareRequest is the [in] parameter set of NetprPathCompare: the [unique]
// server name, the two (ref) path names, the path type, and the comparison flags.
type netprPathCompareRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	PathName1  ndr.WSTR
	PathName2  ndr.WSTR
	PathType   ndr.DWORD
	Flags      ndr.DWORD
}

func (*netprPathCompareRequest) Opnum() uint16 { return srvsvc.OpnumNetprPathCompare }

// netprPathCompareResponse is the reply: the long comparison result is the RPC return
// value (NetprPathCompare returns a signed long, not a NET_API_STATUS).
type netprPathCompareResponse struct {
	Result int32 `ndr:"retval"`
}

// NetprPathCompare calls NetprPathCompare (opnum 32), comparing two path names; it returns
// the signed comparison result (-1, 0, or 1) ([MS-SRVS] 3.1.4.28). The result is the RPC
// return value, so no status check is performed.
func NetprPathCompare(rpc ndr.Invoker, serverName, pathName1, pathName2 string, pathType, flags uint32) (int32, error) {
	req := &netprPathCompareRequest{
		ServerName: optWStr(serverName),
		PathName1:  ndr.WSTR(pathName1),
		PathName2:  ndr.WSTR(pathName2),
		PathType:   ndr.DWORD(pathType),
		Flags:      ndr.DWORD(flags),
	}
	var resp netprPathCompareResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("NetprPathCompare: %w", err)
	}
	return resp.Result, nil
}
