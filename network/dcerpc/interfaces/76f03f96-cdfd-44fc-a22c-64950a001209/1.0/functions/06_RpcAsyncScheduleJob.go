package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncScheduleJobRequest carries the [in] parameters of RpcAsyncScheduleJob.
type rpcAsyncScheduleJobRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	JobId    ndr.DWORD
}

func (*rpcAsyncScheduleJobRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncScheduleJob }

// rpcAsyncScheduleJobResponse carries the [out] parameters and return value of RpcAsyncScheduleJob.
type rpcAsyncScheduleJobResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncScheduleJob calls RpcAsyncScheduleJob (opnum 6) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncScheduleJob(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, jobId ndr.DWORD) (err error) {
	req := &rpcAsyncScheduleJobRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
	}
	var resp rpcAsyncScheduleJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncScheduleJob: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncScheduleJob failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
