package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcScheduleJobRequest carries the [in] parameters of RpcScheduleJob.
type rpcScheduleJobRequest struct {
	HPrinter structures.PRINTER_HANDLE
	JobId    ndr.DWORD
}

func (*rpcScheduleJobRequest) Opnum() uint16 { return winspool.OpnumRpcScheduleJob }

// rpcScheduleJobResponse carries the [out] parameters and return value of RpcScheduleJob.
type rpcScheduleJobResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcScheduleJob calls RpcScheduleJob (opnum 25) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcScheduleJob(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, jobId ndr.DWORD) (err error) {
	req := &rpcScheduleJobRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
	}
	var resp rpcScheduleJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcScheduleJob: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcScheduleJob failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
