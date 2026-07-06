package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// schRpcHighestVersionRequest carries the [in] parameters of SchRpcHighestVersion.
type schRpcHighestVersionRequest struct {
}

func (*schRpcHighestVersionRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcHighestVersion
}

// schRpcHighestVersionResponse carries the [out] parameters and return value of SchRpcHighestVersion.
type schRpcHighestVersionResponse struct {
	PVersion ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// SchRpcHighestVersion calls SchRpcHighestVersion (opnum 0) ([MS-TSCH] section 3.2.5.4.1).
func SchRpcHighestVersion(rpc ndr.Invoker) (PVersion ndr.DWORD, err error) {
	req := &schRpcHighestVersionRequest{}
	var resp schRpcHighestVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcHighestVersion: %w", err)
		return
	}
	PVersion = resp.PVersion
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcHighestVersion failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
