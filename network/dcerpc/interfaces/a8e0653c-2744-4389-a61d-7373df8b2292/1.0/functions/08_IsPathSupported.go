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

// isPathSupportedRequest carries the [in] parameters of IsPathSupported.
type isPathSupportedRequest struct {
	ShareName ndr.WSTR
}

func (*isPathSupportedRequest) Opnum() uint16 { return FileServerVssAgent.OpnumIsPathSupported }

// isPathSupportedResponse carries the [out] parameters and return value of IsPathSupported.
// OwnerMachineName is [out][string] LPWSTR* — a [unique] pointer to a wide string (the
// outer * is the [out] by-reference mechanism and is not transmitted).
type isPathSupportedResponse struct {
	SupportedByThisProvider ndr.BOOL
	OwnerMachineName        *ndr.WSTR `ndr:"unique"`
	Status                  ndr.DWORD `ndr:"retval"`
}

// IsPathSupported calls IsPathSupported (opnum 8) ([MS-FSRVP] — verify the parameter
// modeling and status handling).
func IsPathSupported(rpc ndr.Invoker, shareName ndr.WSTR) (SupportedByThisProvider ndr.BOOL, OwnerMachineName *ndr.WSTR, err error) {
	req := &isPathSupportedRequest{
		ShareName: shareName,
	}
	var resp isPathSupportedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IsPathSupported: %w", err)
		return
	}
	SupportedByThisProvider = resp.SupportedByThisProvider
	OwnerMachineName = resp.OwnerMachineName
	if uint32(resp.Status) != FileServerVssAgent.StatusSuccess {
		err = fmt.Errorf("IsPathSupported failed: %s", FileServerVssAgent.StatusString(uint32(resp.Status)))
	}
	return
}
