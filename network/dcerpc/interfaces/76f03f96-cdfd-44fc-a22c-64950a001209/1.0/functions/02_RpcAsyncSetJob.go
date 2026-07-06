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

// rpcAsyncSetJobRequest carries the [in] parameters of RpcAsyncSetJob.
type rpcAsyncSetJobRequest struct {
	HPrinter      mspar.PRINTER_HANDLE
	JobId         ndr.DWORD
	PJobContainer *mspar.JOB_CONTAINER `ndr:"unique"`
	Command       ndr.DWORD
}

func (*rpcAsyncSetJobRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncSetJob }

// rpcAsyncSetJobResponse carries the [out] parameters and return value of RpcAsyncSetJob.
type rpcAsyncSetJobResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetJob calls RpcAsyncSetJob (opnum 2) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetJob(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD, pJobContainer *mspar.JOB_CONTAINER, command ndr.DWORD) (err error) {
	req := &rpcAsyncSetJobRequest{
		HPrinter:      hPrinter,
		JobId:         jobId,
		PJobContainer: pJobContainer,
		Command:       command,
	}
	var resp rpcAsyncSetJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetJob: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetJob failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
