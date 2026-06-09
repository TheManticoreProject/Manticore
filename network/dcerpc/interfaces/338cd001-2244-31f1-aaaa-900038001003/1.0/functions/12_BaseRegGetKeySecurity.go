package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegGetKeySecurityRequest carries the [in] parameters of BaseRegGetKeySecurity.
type baseRegGetKeySecurityRequest struct {
	HKey                     structures.RPC_HKEY
	SecurityInformation      ndr.DWORD
	PRpcSecurityDescriptorIn structures.RPC_SECURITY_DESCRIPTOR
}

func (*baseRegGetKeySecurityRequest) Opnum() uint16 { return winreg.OpnumBaseRegGetKeySecurity }

// baseRegGetKeySecurityResponse carries the [out] parameters and return value of BaseRegGetKeySecurity.
type baseRegGetKeySecurityResponse struct {
	PRpcSecurityDescriptorOut structures.RPC_SECURITY_DESCRIPTOR
	Status                    ndr.DWORD `ndr:"retval"`
}

// BaseRegGetKeySecurity calls BaseRegGetKeySecurity (opnum 12) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegGetKeySecurity(rpc ndr.Invoker, hKey structures.RPC_HKEY, securityInformation ndr.DWORD, pRpcSecurityDescriptorIn structures.RPC_SECURITY_DESCRIPTOR) (PRpcSecurityDescriptorOut structures.RPC_SECURITY_DESCRIPTOR, err error) {
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
