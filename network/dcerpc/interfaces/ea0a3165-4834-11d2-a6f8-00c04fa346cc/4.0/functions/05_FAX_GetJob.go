package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetJobRequest carries the [in] parameters of FAX_GetJob.
type fAX_GetJobRequest struct {
	JobId ndr.DWORD
}

func (*fAX_GetJobRequest) Opnum() uint16 { return fax.OpnumFAX_GetJob }

// fAX_GetJobResponse carries the [out] parameters and return value of FAX_GetJob.
type fAX_GetJobResponse struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_GetJob calls FAX_GetJob (opnum 5) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetJob(rpc ndr.Invoker, jobId ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, err error) {
	req := &fAX_GetJobRequest{
		JobId: jobId,
	}
	var resp fAX_GetJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetJob: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetJob failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
