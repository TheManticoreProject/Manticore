package ms_rrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// BaseRegGetKeySecurity calls BaseRegGetKeySecurity (opnum 12): returns the parts of a
// key's security descriptor selected by securityInformation (an OWNER/GROUP/DACL/SACL
// bitmask).
func (r *RemoteRegistry) BaseRegGetKeySecurity(hKey structures.RPC_HKEY, securityInformation ndr.DWORD, pRpcSecurityDescriptorIn structures.RPC_SECURITY_DESCRIPTOR) (structures.RPC_SECURITY_DESCRIPTOR, error) {
	if err := r.ensure(); err != nil {
		return structures.RPC_SECURITY_DESCRIPTOR{}, err
	}
	return functions.BaseRegGetKeySecurity(r.rpc, hKey, securityInformation, pRpcSecurityDescriptorIn)
}

// BaseRegSetKeySecurity calls BaseRegSetKeySecurity (opnum 21): sets the parts of a key's
// security descriptor selected by securityInformation.
func (r *RemoteRegistry) BaseRegSetKeySecurity(hKey structures.RPC_HKEY, securityInformation ndr.DWORD, pRpcSecurityDescriptor structures.RPC_SECURITY_DESCRIPTOR) error {
	if err := r.ensure(); err != nil {
		return err
	}
	return functions.BaseRegSetKeySecurity(r.rpc, hKey, securityInformation, pRpcSecurityDescriptor)
}
