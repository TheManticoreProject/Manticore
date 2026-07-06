package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetProtocolStatusRequest carries the [in] parameters of RpcGetProtocolStatus.
type rpcGetProtocolStatusRequest struct {
	SessionId ndr.DWORD
	InfoType  mststs.PROTOCOLSTATUS_INFO_TYPE
}

func (*rpcGetProtocolStatusRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetProtocolStatus }

// rpcGetProtocolStatusResponse carries the [out] parameters and return value of RpcGetProtocolStatus.
type rpcGetProtocolStatusResponse struct {
	PpProtoStatus  []uint8 `ndr:"unique,conformant"`
	PcbProtoStatus ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// RpcGetProtocolStatus calls RpcGetProtocolStatus (opnum 2) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetProtocolStatus(rpc ndr.Invoker, sessionId ndr.DWORD, infoType mststs.PROTOCOLSTATUS_INFO_TYPE) (PpProtoStatus []uint8, PcbProtoStatus ndr.DWORD, err error) {
	req := &rpcGetProtocolStatusRequest{
		SessionId: sessionId,
		InfoType:  infoType,
	}
	var resp rpcGetProtocolStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetProtocolStatus: %w", err)
		return
	}
	PpProtoStatus = resp.PpProtoStatus
	PcbProtoStatus = resp.PcbProtoStatus
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetProtocolStatus failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
