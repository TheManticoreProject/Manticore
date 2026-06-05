package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrShareDelExRequest is the [in] parameter set of NetrShareDelEx: the optional server
// name, the info level, and the inline [in, switch_is(Level)] SHARE_INFO union identifying
// the share to delete (its Tag is set to Level before marshalling).
type netrShareDelExRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
	ShareInfo  structures.SHARE_INFO
}

func (*netrShareDelExRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareDelEx
}

// NetrShareDelEx calls NetrShareDelEx (opnum 57), deleting a share identified by the
// supplied share-info structure ([MS-SRVS] 3.1.4.18). The union discriminant is set to
// level before marshalling.
func NetrShareDelEx(rpc ndr.Invoker, serverName string, level ndr.DWORD, shareInfo structures.SHARE_INFO) error {
	shareInfo.Tag = level
	req := &netrShareDelExRequest{
		ServerName: optWStr(serverName),
		Level:      level,
		ShareInfo:  shareInfo,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrShareDelEx: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrShareDelEx failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return nil
}
