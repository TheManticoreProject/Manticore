package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// setContextRequest carries the [in] parameters of SetContext.
type setContextRequest struct {
	Context ndr.DWORD
}

func (*setContextRequest) Opnum() uint16 { return FileServerVssAgent.OpnumSetContext }

// setContextResponse carries the [out] parameters and return value of SetContext.
type setContextResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SetContext calls SetContext (opnum 1) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func SetContext(rpc ndr.Invoker, context ndr.DWORD) (err error) {
	req := &setContextRequest{
		Context: context,
	}
	var resp setContextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SetContext: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("SetContext failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
