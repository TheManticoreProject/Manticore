package functions

// IDL source: [MS-FASP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fasp/1503b9d7-7fec-4793-9972-6ad58720c9db
// A fetched copy is kept at ms-fasp.idl in the interface directory.

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWSetConfigRequest carries the [in] parameters of RRPC_FWSetConfig.
type rRPC_FWSetConfigRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	ConfigID     msfasp.FW_PROFILE_CONFIG
	Profile      msfasp.FW_PROFILE_TYPE
	PConfig      msfasp.FW_PROFILE_CONFIG_VALUE
	DwBufSize    ndr.DWORD
}

func (*rRPC_FWSetConfigRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWSetConfig }

// rRPC_FWSetConfigResponse carries the [out] parameters and return value of RRPC_FWSetConfig.
type rRPC_FWSetConfigResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetConfig calls RRPC_FWSetConfig (opnum 11) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetConfig(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, configID msfasp.FW_PROFILE_CONFIG, profile msfasp.FW_PROFILE_TYPE, pConfig msfasp.FW_PROFILE_CONFIG_VALUE, dwBufSize ndr.DWORD) (err error) {
	req := &rRPC_FWSetConfigRequest{
		HPolicyStore: hPolicyStore,
		ConfigID:     configID,
		Profile:      profile,
		PConfig:      pConfig,
		DwBufSize:    dwBufSize,
	}
	var resp rRPC_FWSetConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetConfig: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetConfig failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
