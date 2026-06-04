package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// netrDfsModifyPrefixRequest is the [in] parameter set of NetrDfsModifyPrefix: the
// [unique] server name, the inline GUID (a single [in] pointer in the IDL), and the (ref)
// prefix.
type netrDfsModifyPrefixRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Uid        guid.GUID
	Prefix     ndr.WSTR
}

func (*netrDfsModifyPrefixRequest) Opnum() uint16 { return srvsvc.OpnumNetrDfsModifyPrefix }

// NetrDfsModifyPrefix calls NetrDfsModifyPrefix (opnum 50), changing the prefix of a DFS
// entry ([MS-SRVS] 3.1.4.50).
func NetrDfsModifyPrefix(rpc *client.Client, serverName string, uid guid.GUID, prefix string) error {
	req := &netrDfsModifyPrefixRequest{
		ServerName: optWStr(serverName),
		Uid:        uid,
		Prefix:     ndr.WSTR(prefix),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrDfsModifyPrefix: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrDfsModifyPrefix failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
