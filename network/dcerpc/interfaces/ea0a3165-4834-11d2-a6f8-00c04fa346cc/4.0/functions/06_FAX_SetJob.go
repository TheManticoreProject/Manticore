package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetJobRequest carries the [in] parameters of FAX_SetJob.
type fAX_SetJobRequest struct {
	JobId   ndr.DWORD
	Command ndr.DWORD
}

func (*fAX_SetJobRequest) Opnum() uint16 { return fax.OpnumFAX_SetJob }

// fAX_SetJobResponse carries the [out] parameters and return value of FAX_SetJob.
type fAX_SetJobResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetJob calls FAX_SetJob (opnum 6) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetJob(rpc ndr.Invoker, jobId ndr.DWORD, command ndr.DWORD) (err error) {
	req := &fAX_SetJobRequest{
		JobId:   jobId,
		Command: command,
	}
	var resp fAX_SetJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetJob: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetJob failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
