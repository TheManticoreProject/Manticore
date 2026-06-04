package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
)

// netprNameCompareRequest is the [in] parameter set of NetprNameCompare: the [unique]
// server name, the two (ref) names, the name type, and the comparison flags.
type netprNameCompareRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Name1      ndr.WSTR
	Name2      ndr.WSTR
	NameType   ndr.DWORD
	Flags      ndr.DWORD
}

func (*netprNameCompareRequest) Opnum() uint16 { return srvsvc.OpnumNetprNameCompare }

// netprNameCompareResponse is the reply: the long comparison result is the RPC return
// value (NetprNameCompare returns a signed long, not a NET_API_STATUS).
type netprNameCompareResponse struct {
	Result int32 `ndr:"retval"`
}

// NetprNameCompare calls NetprNameCompare (opnum 35), comparing two names; it returns the
// signed comparison result (-1, 0, or 1) ([MS-SRVS] 3.1.4.32). The result is the RPC
// return value, so no status check is performed.
func NetprNameCompare(rpc *client.Client, serverName, name1, name2 string, nameType, flags uint32) (int32, error) {
	req := &netprNameCompareRequest{
		ServerName: optWStr(serverName),
		Name1:      ndr.WSTR(name1),
		Name2:      ndr.WSTR(name2),
		NameType:   ndr.DWORD(nameType),
		Flags:      ndr.DWORD(flags),
	}
	var resp netprNameCompareResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("NetprNameCompare: %w", err)
	}
	return resp.Result, nil
}
