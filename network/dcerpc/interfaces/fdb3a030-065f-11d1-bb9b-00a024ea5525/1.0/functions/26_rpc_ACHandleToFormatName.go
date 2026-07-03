package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACHandleToFormatNameRequest carries the [in] parameters of rpc_ACHandleToFormatName.
type rpc_ACHandleToFormatNameRequest struct {
	HQueue                   msmqmp.RPC_QUEUE_HANDLE
	DwFormatNameRPCBufferLen ndr.DWORD
	LpwcsFormatName          []uint16 `ndr:"unique,size_is=DwFormatNameRPCBufferLen,varying,length_is=DwFormatNameRPCBufferLen"`
	PdwLength                ndr.DWORD
}

func (*rpc_ACHandleToFormatNameRequest) Opnum() uint16 { return qmcomm.Opnumrpc_ACHandleToFormatName }

// rpc_ACHandleToFormatNameResponse carries the [out] parameters and return value of rpc_ACHandleToFormatName.
type rpc_ACHandleToFormatNameResponse struct {
	LpwcsFormatName []uint16 `ndr:"unique,size_is=DwFormatNameRPCBufferLen,varying,length_is=DwFormatNameRPCBufferLen"`
	PdwLength       ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// Rpc_ACHandleToFormatName calls rpc_ACHandleToFormatName (opnum 26) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func Rpc_ACHandleToFormatName(rpc ndr.Invoker, hQueue msmqmp.RPC_QUEUE_HANDLE, dwFormatNameRPCBufferLen ndr.DWORD, lpwcsFormatName []uint16, pdwLength ndr.DWORD) (LpwcsFormatName []uint16, PdwLength ndr.DWORD, err error) {
	req := &rpc_ACHandleToFormatNameRequest{
		HQueue:                   hQueue,
		DwFormatNameRPCBufferLen: dwFormatNameRPCBufferLen,
		LpwcsFormatName:          lpwcsFormatName,
		PdwLength:                pdwLength,
	}
	var resp rpc_ACHandleToFormatNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("rpc_ACHandleToFormatName: %w", err)
		return
	}
	LpwcsFormatName = resp.LpwcsFormatName
	PdwLength = resp.PdwLength
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("rpc_ACHandleToFormatName failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
