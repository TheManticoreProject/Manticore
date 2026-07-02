package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWSetGlobalConfigRequest carries the [in] parameters of RRPC_FWSetGlobalConfig.
type rRPC_FWSetGlobalConfigRequest struct {
	BinaryVersion uint16
	StoreType     msfasp.FW_STORE_TYPE
	ConfigID      msfasp.FW_GLOBAL_CONFIG
	LpBuffer      []uint8 `ndr:"unique,size_is=DwBufSize"`
	DwBufSize     ndr.DWORD
}

func (*rRPC_FWSetGlobalConfigRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWSetGlobalConfig }

// rRPC_FWSetGlobalConfigResponse carries the [out] parameters and return value of RRPC_FWSetGlobalConfig.
type rRPC_FWSetGlobalConfigResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetGlobalConfig calls RRPC_FWSetGlobalConfig (opnum 4) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetGlobalConfig(rpc ndr.Invoker, binaryVersion uint16, storeType msfasp.FW_STORE_TYPE, configID msfasp.FW_GLOBAL_CONFIG, lpBuffer []uint8, dwBufSize ndr.DWORD) (err error) {
	req := &rRPC_FWSetGlobalConfigRequest{
		BinaryVersion: binaryVersion,
		StoreType:     storeType,
		ConfigID:      configID,
		LpBuffer:      lpBuffer,
		DwBufSize:     dwBufSize,
	}
	var resp rRPC_FWSetGlobalConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetGlobalConfig: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetGlobalConfig failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
