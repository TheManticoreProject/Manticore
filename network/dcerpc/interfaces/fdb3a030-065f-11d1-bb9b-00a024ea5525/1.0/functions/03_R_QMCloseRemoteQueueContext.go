package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// r_QMCloseRemoteQueueContextRequest carries the [in, out] context handle of
// R_QMCloseRemoteQueueContext.
type r_QMCloseRemoteQueueContextRequest struct {
	PphContext msmqmp.PCTX_OPENREMOTE_HANDLE_TYPE
}

func (*r_QMCloseRemoteQueueContextRequest) Opnum() uint16 {
	return qmcomm.OpnumR_QMCloseRemoteQueueContext
}

// r_QMCloseRemoteQueueContextResponse carries the [in, out] context handle returned by
// R_QMCloseRemoteQueueContext. The method is declared void in the IDL, so there is no
// trailing HRESULT on the wire — only the (typically nulled) context handle.
type r_QMCloseRemoteQueueContextResponse struct {
	PphContext msmqmp.PCTX_OPENREMOTE_HANDLE_TYPE
}

// R_QMCloseRemoteQueueContext calls R_QMCloseRemoteQueueContext (opnum 3, [MS-MQMP] 3.1.4.3).
// The method returns void; it closes the remote-open context handle and returns the handle
// as the server left it (the server nulls it on success).
func R_QMCloseRemoteQueueContext(rpc ndr.Invoker, pphContext msmqmp.PCTX_OPENREMOTE_HANDLE_TYPE) (PphContext msmqmp.PCTX_OPENREMOTE_HANDLE_TYPE, err error) {
	req := &r_QMCloseRemoteQueueContextRequest{
		PphContext: pphContext,
	}
	var resp r_QMCloseRemoteQueueContextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		return PphContext, fmt.Errorf("R_QMCloseRemoteQueueContext: %w", err)
	}
	return resp.PphContext, nil
}
