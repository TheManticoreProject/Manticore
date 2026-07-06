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
