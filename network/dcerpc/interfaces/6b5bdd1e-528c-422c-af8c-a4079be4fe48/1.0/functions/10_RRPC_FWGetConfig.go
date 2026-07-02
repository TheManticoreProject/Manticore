package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWGetConfigRequest carries the [in] parameters of RRPC_FWGetConfig.
type rRPC_FWGetConfigRequest struct {
	HPolicyStore      msfasp.FW_POLICY_STORE_HANDLE
	ConfigID          msfasp.FW_PROFILE_CONFIG
	Profile           msfasp.FW_PROFILE_TYPE
	DwFlags           ndr.DWORD
	PBuffer           []uint8 `ndr:"unique,size_is=CbData,varying"`
	CbData            ndr.DWORD
	PcbTransmittedLen ndr.DWORD
}

func (*rRPC_FWGetConfigRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWGetConfig }

// rRPC_FWGetConfigResponse carries the [out] parameters and return value of RRPC_FWGetConfig.
type rRPC_FWGetConfigResponse struct {
	PBuffer           []uint8 `ndr:"unique,size_is=CbData,varying"`
	PcbTransmittedLen ndr.DWORD
	PcbRequired       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// RRPC_FWGetConfig calls RRPC_FWGetConfig (opnum 10) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWGetConfig(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, configID msfasp.FW_PROFILE_CONFIG, profile msfasp.FW_PROFILE_TYPE, dwFlags ndr.DWORD, pBuffer []uint8, cbData ndr.DWORD, pcbTransmittedLen ndr.DWORD) (PBuffer []uint8, PcbTransmittedLen ndr.DWORD, PcbRequired ndr.DWORD, err error) {
	req := &rRPC_FWGetConfigRequest{
		HPolicyStore:      hPolicyStore,
		ConfigID:          configID,
		Profile:           profile,
		DwFlags:           dwFlags,
		PBuffer:           pBuffer,
		CbData:            cbData,
		PcbTransmittedLen: pcbTransmittedLen,
	}
	var resp rRPC_FWGetConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWGetConfig: %w", err)
		return
	}
	PBuffer = resp.PBuffer
	PcbTransmittedLen = resp.PcbTransmittedLen
	PcbRequired = resp.PcbRequired
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWGetConfig failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
