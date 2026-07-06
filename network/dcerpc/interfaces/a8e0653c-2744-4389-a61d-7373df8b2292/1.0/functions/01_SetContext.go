package functions

// IDL source: [MS-FSRVP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fsrvp/23382633-78f1-419e-bad0-699dff0c6ef1
// A fetched copy is kept at ms-fsrvp.idl in the interface directory.

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
