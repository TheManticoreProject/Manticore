package functions

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msmqqp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqqp"
)

// remoteQMOpenQueueRequest carries the [in] parameters of RemoteQMOpenQueue.
type remoteQMOpenQueueRequest struct {
	PLicGuid   guid.GUID
	DwMQS      ndr.DWORD
	HQueue     ndr.DWORD
	PQueue     ndr.DWORD
	DwpContext ndr.DWORD
}

func (*remoteQMOpenQueueRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMOpenQueue }

// remoteQMOpenQueueResponse carries the [out] parameters and return value of RemoteQMOpenQueue.
type remoteQMOpenQueueResponse struct {
	PhContext msmqqp.PCTX_RRSESSION_HANDLE_TYPE
	Status    ndr.DWORD `ndr:"retval"`
}

// RemoteQMOpenQueue calls RemoteQMOpenQueue (opnum 2) ([MS-MQQP] 3.1.4.3).
func RemoteQMOpenQueue(rpc ndr.Invoker, pLicGuid guid.GUID, dwMQS ndr.DWORD, hQueue ndr.DWORD, pQueue ndr.DWORD, dwpContext ndr.DWORD) (PhContext msmqqp.PCTX_RRSESSION_HANDLE_TYPE, err error) {
	req := &remoteQMOpenQueueRequest{
		PLicGuid:   pLicGuid,
		DwMQS:      dwMQS,
		HQueue:     hQueue,
		PQueue:     pQueue,
		DwpContext: dwpContext,
	}
	var resp remoteQMOpenQueueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMOpenQueue: %w", err)
		return
	}
	PhContext = resp.PhContext
	if uint32(resp.Status) != qm2qm.StatusSuccess {
		err = fmt.Errorf("RemoteQMOpenQueue failed: %s", qm2qm.StatusString(uint32(resp.Status)))
	}
	return
}
