package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rdcCloseRequest carries the [in] parameters of RdcClose.
type rdcCloseRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
}

func (*rdcCloseRequest) Opnum() uint16 { return FrsTransport.OpnumRdcClose }

// rdcCloseResponse carries the [out] parameters and return value of RdcClose.
type rdcCloseResponse struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
	Status        ndr.DWORD `ndr:"retval"`
}

// RdcClose calls RdcClose (opnum 12) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RdcClose(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT) (ServerContext msfrs2.PFRS_SERVER_CONTEXT, err error) {
	req := &rdcCloseRequest{
		ServerContext: serverContext,
	}
	var resp rdcCloseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RdcClose: %w", err)
		return
	}
	ServerContext = resp.ServerContext
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RdcClose failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
