package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMOpenRemoteQueueRequest carries the [in] parameters of R_QMOpenRemoteQueue.
type r_QMOpenRemoteQueueRequest struct {
	PQueueFormat       *msmqmq.QUEUE_FORMAT `ndr:"unique"`
	DwCallingProcessID ndr.DWORD
	DwDesiredAccess    ndr.DWORD
	DwShareMode        ndr.DWORD
	PLicGuid           dtyp.GUID
	DwMQS              ndr.DWORD
}

func (*r_QMOpenRemoteQueueRequest) Opnum() uint16 { return qmcomm.OpnumR_QMOpenRemoteQueue }

// r_QMOpenRemoteQueueResponse carries the [out] parameters and return value of R_QMOpenRemoteQueue.
type r_QMOpenRemoteQueueResponse struct {
	PphContext msmqmp.PCTX_OPENREMOTE_HANDLE_TYPE
	PdwContext ndr.DWORD
	DwpQueue   ndr.DWORD
	PhQueue    ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// R_QMOpenRemoteQueue calls R_QMOpenRemoteQueue (opnum 2) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMOpenRemoteQueue(rpc ndr.Invoker, pQueueFormat *msmqmq.QUEUE_FORMAT, dwCallingProcessID ndr.DWORD, dwDesiredAccess ndr.DWORD, dwShareMode ndr.DWORD, pLicGuid dtyp.GUID, dwMQS ndr.DWORD) (PphContext msmqmp.PCTX_OPENREMOTE_HANDLE_TYPE, PdwContext ndr.DWORD, DwpQueue ndr.DWORD, PhQueue ndr.DWORD, err error) {
	req := &r_QMOpenRemoteQueueRequest{
		PQueueFormat:       pQueueFormat,
		DwCallingProcessID: dwCallingProcessID,
		DwDesiredAccess:    dwDesiredAccess,
		DwShareMode:        dwShareMode,
		PLicGuid:           pLicGuid,
		DwMQS:              dwMQS,
	}
	var resp r_QMOpenRemoteQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMOpenRemoteQueue: %w", err)
		return
	}
	PphContext = resp.PphContext
	PdwContext = resp.PdwContext
	DwpQueue = resp.DwpQueue
	PhQueue = resp.PhQueue
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMOpenRemoteQueue failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
