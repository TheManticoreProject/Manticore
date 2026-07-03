package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// rpc_QMOpenQueueInternalRequest carries the [in] parameters of rpc_QMOpenQueueInternal.
type rpc_QMOpenQueueInternalRequest struct {
	PQueueFormat        msmqmq.QUEUE_FORMAT
	DwDesiredAccess     ndr.DWORD
	DwShareMode         ndr.DWORD
	HRemoteQueue        ndr.DWORD
	LplpRemoteQueueName *ndr.WSTR `ndr:"ptr"`
	DwpQueue            ndr.DWORD
	PLicGuid            dtyp.GUID
	LpClientName        ndr.WSTR
	DwRemoteProtocol    ndr.DWORD
	DwpRemoteContext    ndr.DWORD
}

func (*rpc_QMOpenQueueInternalRequest) Opnum() uint16 { return qmcomm.Opnumrpc_QMOpenQueueInternal }

// rpc_QMOpenQueueInternalResponse carries the [out] parameters and return value of rpc_QMOpenQueueInternal.
type rpc_QMOpenQueueInternalResponse struct {
	LplpRemoteQueueName *ndr.WSTR `ndr:"ptr"`
	PdwQMContext        ndr.DWORD
	PhQueue             msmqmp.RPC_QUEUE_HANDLE
	Status              ndr.DWORD `ndr:"retval"`
}

// Rpc_QMOpenQueueInternal calls rpc_QMOpenQueueInternal (opnum 19) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func Rpc_QMOpenQueueInternal(rpc ndr.Invoker, pQueueFormat msmqmq.QUEUE_FORMAT, dwDesiredAccess ndr.DWORD, dwShareMode ndr.DWORD, hRemoteQueue ndr.DWORD, lplpRemoteQueueName *ndr.WSTR, dwpQueue ndr.DWORD, pLicGuid dtyp.GUID, lpClientName ndr.WSTR, dwRemoteProtocol ndr.DWORD, dwpRemoteContext ndr.DWORD) (LplpRemoteQueueName *ndr.WSTR, PdwQMContext ndr.DWORD, PhQueue msmqmp.RPC_QUEUE_HANDLE, err error) {
	req := &rpc_QMOpenQueueInternalRequest{
		PQueueFormat:        pQueueFormat,
		DwDesiredAccess:     dwDesiredAccess,
		DwShareMode:         dwShareMode,
		HRemoteQueue:        hRemoteQueue,
		LplpRemoteQueueName: lplpRemoteQueueName,
		DwpQueue:            dwpQueue,
		PLicGuid:            pLicGuid,
		LpClientName:        lpClientName,
		DwRemoteProtocol:    dwRemoteProtocol,
		DwpRemoteContext:    dwpRemoteContext,
	}
	var resp rpc_QMOpenQueueInternalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_QMOpenQueueInternal: %w", err)
		return
	}
	LplpRemoteQueueName = resp.LplpRemoteQueueName
	PdwQMContext = resp.PdwQMContext
	PhQueue = resp.PhQueue
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("rpc_QMOpenQueueInternal failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
