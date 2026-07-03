package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcSetJobRequest carries the [in] parameters of RpcSetJob.
type rpcSetJobRequest struct {
	HPrinter      msrprn.PRINTER_HANDLE
	JobId         ndr.DWORD
	PJobContainer *msrprn.JOB_CONTAINER `ndr:"unique"`
	Command       ndr.DWORD
}

func (*rpcSetJobRequest) Opnum() uint16 { return winspool.OpnumRpcSetJob }

// rpcSetJobResponse carries the [out] parameters and return value of RpcSetJob.
type rpcSetJobResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetJob calls RpcSetJob (opnum 2) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetJob(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, pJobContainer *msrprn.JOB_CONTAINER, command ndr.DWORD) (err error) {
	req := &rpcSetJobRequest{
		HPrinter:      hPrinter,
		JobId:         jobId,
		PJobContainer: pJobContainer,
		Command:       command,
	}
	var resp rpcSetJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetJob: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetJob failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
