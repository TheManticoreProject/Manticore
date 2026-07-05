package functions

import (
	"fmt"

	InitShutdown "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/894de0c0-0d55-11d3-a322-00c04fa321a1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseAbortShutdownRequest carries the [in] parameters of BaseAbortShutdown.
type baseAbortShutdownRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*baseAbortShutdownRequest) Opnum() uint16 { return InitShutdown.OpnumBaseAbortShutdown }

// baseAbortShutdownResponse carries the [out] parameters and return value of BaseAbortShutdown.
type baseAbortShutdownResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseAbortShutdown calls BaseAbortShutdown (opnum 1) ([MS-RSP] section 3.2.4.2). It stops a system shutdown that was
// previously requested with BaseInitiateShutdown or BaseInitiateShutdownEx.
func BaseAbortShutdown(rpc ndr.Invoker, serverName *ndr.WSTR) (err error) {
	req := &baseAbortShutdownRequest{
		ServerName: serverName,
	}
	var resp baseAbortShutdownResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseAbortShutdown: %w", err)
		return
	}
	if uint32(resp.Status) != InitShutdown.StatusSuccess {
		err = fmt.Errorf("BaseAbortShutdown failed: %s", InitShutdown.StatusString(uint32(resp.Status)))
	}
	return
}
