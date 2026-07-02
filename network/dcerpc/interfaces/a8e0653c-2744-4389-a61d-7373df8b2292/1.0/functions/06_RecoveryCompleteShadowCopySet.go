package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// recoveryCompleteShadowCopySetRequest carries the [in] parameters of RecoveryCompleteShadowCopySet.
type recoveryCompleteShadowCopySetRequest struct {
	ShadowCopySetId dtyp.GUID
}

func (*recoveryCompleteShadowCopySetRequest) Opnum() uint16 {
	return FileServerVssAgent.OpnumRecoveryCompleteShadowCopySet
}

// recoveryCompleteShadowCopySetResponse carries the [out] parameters and return value of RecoveryCompleteShadowCopySet.
type recoveryCompleteShadowCopySetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RecoveryCompleteShadowCopySet calls RecoveryCompleteShadowCopySet (opnum 6) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func RecoveryCompleteShadowCopySet(rpc ndr.Invoker, shadowCopySetId guid.GUID) (err error) {
	req := &recoveryCompleteShadowCopySetRequest{
		ShadowCopySetId: dtyp.NewGUID(shadowCopySetId),
	}
	var resp recoveryCompleteShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RecoveryCompleteShadowCopySet: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("RecoveryCompleteShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
