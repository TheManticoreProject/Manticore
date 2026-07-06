package functions

// IDL source: [MS-EFSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
// A fetched copy is kept at ms-efsr.idl in the interface directory.

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msefsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-efsr"
)

// efsRpcQueryRecoveryAgentsRequest carries the [in] parameters of EfsRpcQueryRecoveryAgents.
type efsRpcQueryRecoveryAgentsRequest struct {
	FileName ndr.WSTR
}

func (*efsRpcQueryRecoveryAgentsRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcQueryRecoveryAgents }

// efsRpcQueryRecoveryAgentsResponse carries the [out] parameters and return value of EfsRpcQueryRecoveryAgents.
type efsRpcQueryRecoveryAgentsResponse struct {
	RecoveryAgents *msefsr.ENCRYPTION_CERTIFICATE_HASH_LIST `ndr:"unique"`
	Status         ndr.DWORD                                `ndr:"retval"`
}

// EfsRpcQueryRecoveryAgents calls EfsRpcQueryRecoveryAgents (opnum 7) ([MS-EFSR] — verify the parameter
// modeling and status handling).
func EfsRpcQueryRecoveryAgents(rpc ndr.Invoker, fileName ndr.WSTR) (RecoveryAgents *msefsr.ENCRYPTION_CERTIFICATE_HASH_LIST, err error) {
	req := &efsRpcQueryRecoveryAgentsRequest{
		FileName: fileName,
	}
	var resp efsRpcQueryRecoveryAgentsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EfsRpcQueryRecoveryAgents: %w", err)
		return
	}
	RecoveryAgents = resp.RecoveryAgents
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		err = fmt.Errorf("EfsRpcQueryRecoveryAgents failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
