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

// rRPC_FWGetConfig2_10Request carries the [in] parameters of RRPC_FWGetConfig2_10.
type rRPC_FWGetConfig2_10Request struct {
	HPolicyStore      msfasp.FW_POLICY_STORE_HANDLE
	ConfigID          msfasp.FW_PROFILE_CONFIG
	Profile           msfasp.FW_PROFILE_TYPE
	DwFlags           ndr.DWORD
	PBuffer           []uint8 `ndr:"unique,size_is=CbData,varying"`
	CbData            ndr.DWORD
	PcbTransmittedLen ndr.DWORD
}

func (*rRPC_FWGetConfig2_10Request) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWGetConfig2_10 }

// rRPC_FWGetConfig2_10Response carries the [out] parameters and return value of RRPC_FWGetConfig2_10.
type rRPC_FWGetConfig2_10Response struct {
	PBuffer           []uint8 `ndr:"unique,size_is=CbData,varying"`
	PcbTransmittedLen ndr.DWORD
	PcbRequired       ndr.DWORD
	POrigin           msfasp.FW_RULE_ORIGIN_TYPE
	Status            ndr.DWORD `ndr:"retval"`
}

// RRPC_FWGetConfig2_10 calls RRPC_FWGetConfig2_10 (opnum 45) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWGetConfig2_10(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, configID msfasp.FW_PROFILE_CONFIG, profile msfasp.FW_PROFILE_TYPE, dwFlags ndr.DWORD, pBuffer []uint8, cbData ndr.DWORD, pcbTransmittedLen ndr.DWORD) (PBuffer []uint8, PcbTransmittedLen ndr.DWORD, PcbRequired ndr.DWORD, POrigin msfasp.FW_RULE_ORIGIN_TYPE, err error) {
	req := &rRPC_FWGetConfig2_10Request{
		HPolicyStore:      hPolicyStore,
		ConfigID:          configID,
		Profile:           profile,
		DwFlags:           dwFlags,
		PBuffer:           pBuffer,
		CbData:            cbData,
		PcbTransmittedLen: pcbTransmittedLen,
	}
	var resp rRPC_FWGetConfig2_10Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWGetConfig2_10: %w", err)
		return
	}
	PBuffer = resp.PBuffer
	PcbTransmittedLen = resp.PcbTransmittedLen
	PcbRequired = resp.PcbRequired
	POrigin = resp.POrigin
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWGetConfig2_10 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
