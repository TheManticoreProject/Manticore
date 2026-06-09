package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegSetKeySecurityRequest carries the [in] parameters of BaseRegSetKeySecurity.
type baseRegSetKeySecurityRequest struct {
	HKey                   structures.RPC_HKEY
	SecurityInformation    ndr.DWORD
	PRpcSecurityDescriptor structures.RPC_SECURITY_DESCRIPTOR
}

func (*baseRegSetKeySecurityRequest) Opnum() uint16 { return winreg.OpnumBaseRegSetKeySecurity }

// baseRegSetKeySecurityResponse carries the [out] parameters and return value of BaseRegSetKeySecurity.
type baseRegSetKeySecurityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegSetKeySecurity calls BaseRegSetKeySecurity (opnum 21) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegSetKeySecurity(rpc ndr.Invoker, hKey structures.RPC_HKEY, securityInformation ndr.DWORD, pRpcSecurityDescriptor structures.RPC_SECURITY_DESCRIPTOR) (err error) {
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
