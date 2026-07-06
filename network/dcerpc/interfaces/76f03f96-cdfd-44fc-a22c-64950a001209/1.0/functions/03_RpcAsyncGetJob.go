package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetJobRequest carries the [in] parameters of RpcAsyncGetJob.
type rpcAsyncGetJobRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	JobId    ndr.DWORD
	Level    ndr.DWORD
	PJob     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAsyncGetJobRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncGetJob }

// rpcAsyncGetJobResponse carries the [out] parameters and return value of RpcAsyncGetJob.
type rpcAsyncGetJobResponse struct {
	PJob      []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetJob calls RpcAsyncGetJob (opnum 3) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetJob(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD, level ndr.DWORD, pJob []uint8, cbBuf ndr.DWORD) (PJob []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetJobRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
		Level:    level,
		PJob:     pJob,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncGetJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetJob: %w", err)
		return
	}
	PJob = resp.PJob
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetJob failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
