package functions

import (
	"fmt"

	FileServerVssAgent "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/a8e0653c-2744-4389-a61d-7373df8b2292/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// isPathShadowCopiedRequest carries the [in] parameters of IsPathShadowCopied.
type isPathShadowCopiedRequest struct {
	ShareName ndr.WSTR
}

func (*isPathShadowCopiedRequest) Opnum() uint16 { return FileServerVssAgent.OpnumIsPathShadowCopied }

// isPathShadowCopiedResponse carries the [out] parameters and return value of IsPathShadowCopied.
type isPathShadowCopiedResponse struct {
	ShadowCopyPresent       ndr.BOOL
	ShadowCopyCompatibility int32
	Status                  ndr.DWORD `ndr:"retval"`
}

// IsPathShadowCopied calls IsPathShadowCopied (opnum 9) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func IsPathShadowCopied(rpc ndr.Invoker, shareName ndr.WSTR) (ShadowCopyPresent ndr.BOOL, ShadowCopyCompatibility int32, err error) {
	req := &isPathShadowCopiedRequest{
		ShareName: shareName,
	}
	var resp isPathShadowCopiedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IsPathShadowCopied: %w", err)
		return
	}
	ShadowCopyPresent = resp.ShadowCopyPresent
	ShadowCopyCompatibility = resp.ShadowCopyCompatibility
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("IsPathShadowCopied failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
