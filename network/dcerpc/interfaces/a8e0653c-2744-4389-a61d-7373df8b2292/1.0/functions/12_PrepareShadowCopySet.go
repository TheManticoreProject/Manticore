package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// prepareShadowCopySetRequest carries the [in] parameters of PrepareShadowCopySet.
type prepareShadowCopySetRequest struct {
	ShadowCopySetId       dtyp.GUID
	TimeOutInMilliseconds ndr.DWORD
}

func (*prepareShadowCopySetRequest) Opnum() uint16 {
	return FileServerVssAgent.OpnumPrepareShadowCopySet
}

// prepareShadowCopySetResponse carries the [out] parameters and return value of PrepareShadowCopySet.
type prepareShadowCopySetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// PrepareShadowCopySet calls PrepareShadowCopySet (opnum 12) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func PrepareShadowCopySet(rpc ndr.Invoker, shadowCopySetId guid.GUID, timeOutInMilliseconds ndr.DWORD) (err error) {
	req := &prepareShadowCopySetRequest{
		ShadowCopySetId:       dtyp.NewGUID(shadowCopySetId),
		TimeOutInMilliseconds: timeOutInMilliseconds,
	}
	var resp prepareShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PrepareShadowCopySet: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("PrepareShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
