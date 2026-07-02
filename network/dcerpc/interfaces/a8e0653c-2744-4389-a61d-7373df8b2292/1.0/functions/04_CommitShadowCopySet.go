package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// commitShadowCopySetRequest carries the [in] parameters of CommitShadowCopySet.
type commitShadowCopySetRequest struct {
	ShadowCopySetId       dtyp.GUID
	TimeOutInMilliseconds ndr.DWORD
}

func (*commitShadowCopySetRequest) Opnum() uint16 { return FileServerVssAgent.OpnumCommitShadowCopySet }

// commitShadowCopySetResponse carries the [out] parameters and return value of CommitShadowCopySet.
type commitShadowCopySetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// CommitShadowCopySet calls CommitShadowCopySet (opnum 4) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func CommitShadowCopySet(rpc ndr.Invoker, shadowCopySetId guid.GUID, timeOutInMilliseconds ndr.DWORD) (err error) {
	req := &commitShadowCopySetRequest{
		ShadowCopySetId:       dtyp.NewGUID(shadowCopySetId),
		TimeOutInMilliseconds: timeOutInMilliseconds,
	}
	var resp commitShadowCopySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("CommitShadowCopySet: %w", err)
		return
	}
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("CommitShadowCopySet failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
