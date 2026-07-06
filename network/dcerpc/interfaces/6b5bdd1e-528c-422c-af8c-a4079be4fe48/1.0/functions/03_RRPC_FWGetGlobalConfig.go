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

// rRPC_FWGetGlobalConfigRequest carries the [in] parameters of RRPC_FWGetGlobalConfig.
type rRPC_FWGetGlobalConfigRequest struct {
	BinaryVersion     uint16
	StoreType         msfasp.FW_STORE_TYPE
	ConfigID          msfasp.FW_GLOBAL_CONFIG
	DwFlags           ndr.DWORD
	PBuffer           []uint8 `ndr:"unique,size_is=CbData,varying"`
	CbData            ndr.DWORD
	PcbTransmittedLen ndr.DWORD
}

func (*rRPC_FWGetGlobalConfigRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWGetGlobalConfig }

// rRPC_FWGetGlobalConfigResponse carries the [out] parameters and return value of RRPC_FWGetGlobalConfig.
type rRPC_FWGetGlobalConfigResponse struct {
	PBuffer           []uint8 `ndr:"unique,size_is=CbData,varying"`
	PcbTransmittedLen ndr.DWORD
	PcbRequired       ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// RRPC_FWGetGlobalConfig calls RRPC_FWGetGlobalConfig (opnum 3) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWGetGlobalConfig(rpc ndr.Invoker, binaryVersion uint16, storeType msfasp.FW_STORE_TYPE, configID msfasp.FW_GLOBAL_CONFIG, dwFlags ndr.DWORD, pBuffer []uint8, cbData ndr.DWORD, pcbTransmittedLen ndr.DWORD) (PBuffer []uint8, PcbTransmittedLen ndr.DWORD, PcbRequired ndr.DWORD, err error) {
	req := &rRPC_FWGetGlobalConfigRequest{
		BinaryVersion:     binaryVersion,
		StoreType:         storeType,
		ConfigID:          configID,
		DwFlags:           dwFlags,
		PBuffer:           pBuffer,
		CbData:            cbData,
		PcbTransmittedLen: pcbTransmittedLen,
	}
	var resp rRPC_FWGetGlobalConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWGetGlobalConfig: %w", err)
		return
	}
	PBuffer = resp.PBuffer
	PcbTransmittedLen = resp.PcbTransmittedLen
	PcbRequired = resp.PcbRequired
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWGetGlobalConfig failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
