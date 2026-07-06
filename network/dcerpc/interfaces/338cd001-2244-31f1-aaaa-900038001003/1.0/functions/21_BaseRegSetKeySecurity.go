package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegSetKeySecurityRequest carries the [in] parameters of BaseRegSetKeySecurity.
type baseRegSetKeySecurityRequest struct {
	HKey                   msrrp.RPC_HKEY
	SecurityInformation    ndr.DWORD
	PRpcSecurityDescriptor msrrp.RPC_SECURITY_DESCRIPTOR
}

func (*baseRegSetKeySecurityRequest) Opnum() uint16 { return winreg.OpnumBaseRegSetKeySecurity }

// baseRegSetKeySecurityResponse carries the [out] parameters and return value of BaseRegSetKeySecurity.
type baseRegSetKeySecurityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegSetKeySecurity calls BaseRegSetKeySecurity (opnum 21) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegSetKeySecurity(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, securityInformation ndr.DWORD, pRpcSecurityDescriptor msrrp.RPC_SECURITY_DESCRIPTOR) (err error) {
	req := &baseRegSetKeySecurityRequest{
		HKey:                   hKey,
		SecurityInformation:    securityInformation,
		PRpcSecurityDescriptor: pRpcSecurityDescriptor,
	}
	var resp baseRegSetKeySecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegSetKeySecurity: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegSetKeySecurity failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
