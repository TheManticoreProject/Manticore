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

// schRpcGetSecurityRequest carries the [in] parameters of SchRpcGetSecurity.
type schRpcGetSecurityRequest struct {
	Path                ndr.WSTR
	SecurityInformation ndr.DWORD
}

func (*schRpcGetSecurityRequest) Opnum() uint16 { return schrpc.OpnumSchRpcGetSecurity }

// schRpcGetSecurityResponse carries the [out] parameters and return value of SchRpcGetSecurity.
type schRpcGetSecurityResponse struct {
	Sddl   *ndr.WSTR `ndr:"unique"`
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcGetSecurity calls SchRpcGetSecurity (opnum 5) ([MS-TSCH] section 3.2.5.4.6).
func SchRpcGetSecurity(rpc ndr.Invoker, path ndr.WSTR, securityInformation ndr.DWORD) (Sddl *ndr.WSTR, err error) {
	req := &schRpcGetSecurityRequest{
		Path:                path,
		SecurityInformation: securityInformation,
	}
	var resp schRpcGetSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcGetSecurity: %w", err)
		return
	}
	Sddl = resp.Sddl
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcGetSecurity failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
