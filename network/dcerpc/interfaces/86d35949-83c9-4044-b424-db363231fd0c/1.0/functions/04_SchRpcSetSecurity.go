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

// schRpcSetSecurityRequest carries the [in] parameters of SchRpcSetSecurity.
type schRpcSetSecurityRequest struct {
	Path  ndr.WSTR
	Sddl  ndr.WSTR
	Flags ndr.DWORD
}

func (*schRpcSetSecurityRequest) Opnum() uint16 { return schrpc.OpnumSchRpcSetSecurity }

// schRpcSetSecurityResponse carries the [out] parameters and return value of SchRpcSetSecurity.
type schRpcSetSecurityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcSetSecurity calls SchRpcSetSecurity (opnum 4) ([MS-TSCH] section 3.2.5.4.5).
func SchRpcSetSecurity(rpc ndr.Invoker, path ndr.WSTR, sddl ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &schRpcSetSecurityRequest{
		Path:  path,
		Sddl:  sddl,
		Flags: flags,
	}
	var resp schRpcSetSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcSetSecurity: %w", err)
		return
	}
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcSetSecurity failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
