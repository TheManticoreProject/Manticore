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

// rpcAsyncAddJobRequest carries the [in] parameters of RpcAsyncAddJob.
type rpcAsyncAddJobRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	Level    ndr.DWORD
	PAddJob  []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAsyncAddJobRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncAddJob }

// rpcAsyncAddJobResponse carries the [out] parameters and return value of RpcAsyncAddJob.
type rpcAsyncAddJobResponse struct {
	PAddJob   []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddJob calls RpcAsyncAddJob (opnum 5) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddJob(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, level ndr.DWORD, pAddJob []uint8, cbBuf ndr.DWORD) (PAddJob []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncAddJobRequest{
		HPrinter: hPrinter,
		Level:    level,
		PAddJob:  pAddJob,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncAddJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddJob: %w", err)
		return
	}
	PAddJob = resp.PAddJob
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddJob failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
