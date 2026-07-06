package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// abortShadowCopySetRequest carries the [in] parameters of AbortShadowCopySet.
type abortShadowCopySetRequest struct {
	ShadowCopySetId msdtyp.GUID
}

func (*abortShadowCopySetRequest) Opnum() uint16 { return FileServerVssAgent.OpnumAbortShadowCopySet }

// abortShadowCopySetResponse carries the [out] parameters and return value of AbortShadowCopySet.
type abortShadowCopySetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// AbortShadowCopySet calls AbortShadowCopySet (opnum 7) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func AbortShadowCopySet(rpc ndr.Invoker, shadowCopySetId guid.GUID) (err error) {
	req := &abortShadowCopySetRequest{
		ShadowCopySetId: msdtyp.NewGUID(shadowCopySetId),
	}
	var resp abortShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AbortShadowCopySet: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("AbortShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
