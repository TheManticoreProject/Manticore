package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// netrDfsDeleteExitPointRequest is the [in] parameter set of NetrDfsDeleteExitPoint: the
// [unique] server name, the inline GUID (a single [in] pointer in the IDL), the (ref)
// prefix, and the exit-point type.
type netrDfsDeleteExitPointRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Uid        dtyp.GUID
	Prefix     ndr.WSTR
	Type       ndr.DWORD
}

func (*netrDfsDeleteExitPointRequest) Opnum() uint16 { return srvsvc.OpnumNetrDfsDeleteExitPoint }

// NetrDfsDeleteExitPoint calls NetrDfsDeleteExitPoint (opnum 49), deleting a DFS exit
// point ([MS-SRVS] 3.1.4.49).
func NetrDfsDeleteExitPoint(rpc ndr.Invoker, serverName string, uid guid.GUID, prefix string, typ uint32) error {
	req := &netrDfsDeleteExitPointRequest{
		ServerName: optWStr(serverName),
		Uid:        dtyp.NewGUID(uid),
		Prefix:     ndr.WSTR(prefix),
		Type:       ndr.DWORD(typ),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrDfsDeleteExitPoint: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrDfsDeleteExitPoint failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
