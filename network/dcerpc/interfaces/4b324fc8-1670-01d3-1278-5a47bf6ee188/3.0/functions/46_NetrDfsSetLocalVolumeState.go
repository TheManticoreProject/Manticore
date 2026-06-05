package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// netrDfsSetLocalVolumeStateRequest is the [in] parameter set of
// NetrDfsSetLocalVolumeState: the [unique] server name, the inline volume GUID (a single
// [in] pointer in the IDL), the (ref) prefix, and the new state.
type netrDfsSetLocalVolumeStateRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Uid        guid.GUID
	Prefix     ndr.WSTR
	State      ndr.DWORD
}

func (*netrDfsSetLocalVolumeStateRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrDfsSetLocalVolumeState
}

// NetrDfsSetLocalVolumeState calls NetrDfsSetLocalVolumeState (opnum 46), setting the
// state of a local DFS volume ([MS-SRVS] 3.1.4.47).
func NetrDfsSetLocalVolumeState(rpc ndr.Invoker, serverName string, uid guid.GUID, prefix string, state uint32) error {
	req := &netrDfsSetLocalVolumeStateRequest{
		ServerName: optWStr(serverName),
		Uid:        uid,
		Prefix:     ndr.WSTR(prefix),
		State:      ndr.DWORD(state),
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("NetrDfsSetLocalVolumeState: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return fmt.Errorf("NetrDfsSetLocalVolumeState failed: %s", srvsvc.StatusString(status))
	}
	return nil
}
