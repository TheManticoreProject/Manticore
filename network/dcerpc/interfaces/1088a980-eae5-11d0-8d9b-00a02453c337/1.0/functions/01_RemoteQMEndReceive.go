package functions

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqqp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqqp"
)

// remoteQMEndReceiveRequest carries the [in] parameters of RemoteQMEndReceive.
type remoteQMEndReceiveRequest struct {
	PphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE
	DwAck      ndr.DWORD
}

func (*remoteQMEndReceiveRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMEndReceive }

// remoteQMEndReceiveResponse carries the [out] parameters and return value of RemoteQMEndReceive.
type remoteQMEndReceiveResponse struct {
	PphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE
	Status     ndr.DWORD `ndr:"retval"`
}

// RemoteQMEndReceive calls RemoteQMEndReceive (opnum 1) ([MS-MQQP] 3.1.4.2).
func RemoteQMEndReceive(rpc ndr.Invoker, pphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE, dwAck ndr.DWORD) (PphContext msmqqp.PCTX_REMOTEREAD_HANDLE_TYPE, err error) {
	req := &remoteQMEndReceiveRequest{
		PphContext: pphContext,
		DwAck:      dwAck,
	}
	var resp remoteQMEndReceiveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMEndReceive: %w", err)
		return
	}
	PphContext = resp.PphContext
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMEndReceive failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
