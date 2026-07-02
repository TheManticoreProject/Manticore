package functions

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// rdcFileDataTransferKeepAliveRequest carries the [in] parameters of RdcFileDataTransferKeepAlive.
type rdcFileDataTransferKeepAliveRequest struct {
	ServerContext msfrs2.PFRS_SERVER_CONTEXT
}

func (*rdcFileDataTransferKeepAliveRequest) Opnum() uint16 {
	return FrsTransport.OpnumRdcFileDataTransferKeepAlive
}

// rdcFileDataTransferKeepAliveResponse carries the [out] parameters and return value of RdcFileDataTransferKeepAlive.
type rdcFileDataTransferKeepAliveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RdcFileDataTransferKeepAlive calls RdcFileDataTransferKeepAlive (opnum 17) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func RdcFileDataTransferKeepAlive(rpc ndr.Invoker, serverContext msfrs2.PFRS_SERVER_CONTEXT) (err error) {
	req := &rdcFileDataTransferKeepAliveRequest{
		ServerContext: serverContext,
	}
	var resp rdcFileDataTransferKeepAliveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RdcFileDataTransferKeepAlive: %w", err)
		return
	}
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("RdcFileDataTransferKeepAlive failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
