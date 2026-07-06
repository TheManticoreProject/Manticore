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

// baseRegGetKeySecurityRequest carries the [in] parameters of BaseRegGetKeySecurity.
type baseRegGetKeySecurityRequest struct {
	HKey                     msrrp.RPC_HKEY
	SecurityInformation      ndr.DWORD
	PRpcSecurityDescriptorIn msrrp.RPC_SECURITY_DESCRIPTOR
}

func (*baseRegGetKeySecurityRequest) Opnum() uint16 { return winreg.OpnumBaseRegGetKeySecurity }

// baseRegGetKeySecurityResponse carries the [out] parameters and return value of BaseRegGetKeySecurity.
type baseRegGetKeySecurityResponse struct {
	PRpcSecurityDescriptorOut msrrp.RPC_SECURITY_DESCRIPTOR
	Status                    ndr.DWORD `ndr:"retval"`
}

// BaseRegGetKeySecurity calls BaseRegGetKeySecurity (opnum 12) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegGetKeySecurity(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, securityInformation ndr.DWORD, pRpcSecurityDescriptorIn msrrp.RPC_SECURITY_DESCRIPTOR) (PRpcSecurityDescriptorOut msrrp.RPC_SECURITY_DESCRIPTOR, err error) {
	req := &baseRegGetKeySecurityRequest{
		HKey:                     hKey,
		SecurityInformation:      securityInformation,
		PRpcSecurityDescriptorIn: pRpcSecurityDescriptorIn,
	}
	var resp baseRegGetKeySecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegGetKeySecurity: %w", err)
		return
	}
	PRpcSecurityDescriptorOut = resp.PRpcSecurityDescriptorOut
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegGetKeySecurity failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
