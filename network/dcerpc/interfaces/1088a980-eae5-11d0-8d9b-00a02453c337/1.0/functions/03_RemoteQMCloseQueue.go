package functions

// IDL source: [MS-MQQP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqqp/e3ad0b4f-51ab-4a7c-936b-c4f3e6f57b2d
// A fetched copy is kept at ms-mqqp.idl in the interface directory.

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqqp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqqp"
)

// remoteQMCloseQueueRequest carries the [in] parameters of RemoteQMCloseQueue.
type remoteQMCloseQueueRequest struct {
	PphContext msmqqp.PCTX_RRSESSION_HANDLE_TYPE
}

func (*remoteQMCloseQueueRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMCloseQueue }

// remoteQMCloseQueueResponse carries the [out] parameters and return value of RemoteQMCloseQueue.
type remoteQMCloseQueueResponse struct {
	PphContext msmqqp.PCTX_RRSESSION_HANDLE_TYPE
	Status     ndr.DWORD `ndr:"retval"`
}

// RemoteQMCloseQueue calls RemoteQMCloseQueue (opnum 3) ([MS-MQQP] 3.1.4.4).
func RemoteQMCloseQueue(rpc ndr.Invoker, pphContext msmqqp.PCTX_RRSESSION_HANDLE_TYPE) (PphContext msmqqp.PCTX_RRSESSION_HANDLE_TYPE, err error) {
	req := &remoteQMCloseQueueRequest{
		PphContext: pphContext,
	}
	var resp remoteQMCloseQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMCloseQueue: %w", err)
		return
	}
	PphContext = resp.PphContext
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMCloseQueue failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
