package functions

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcQuerySessionDataRequest carries the [in] parameters of RpcQuerySessionData.
type rpcQuerySessionDataRequest struct {
	SessionId     ndr.DWORD
	Type          mststs.QUERY_SESSION_DATA_TYPE
	PbInputData   []uint8 `ndr:"ref,size_is=CbInputData"`
	CbInputData   ndr.DWORD
	CbSessionData ndr.DWORD
}

func (*rpcQuerySessionDataRequest) Opnum() uint16 { return RCMPublic.OpnumRpcQuerySessionData }

// rpcQuerySessionDataResponse carries the [out] parameters and return value of RpcQuerySessionData.
type rpcQuerySessionDataResponse struct {
	PbSessionData        []uint8 `ndr:"ref,size_is=CbSessionData,varying"`
	PcbReturnLength      ndr.DWORD
	PcbRequireBufferSize ndr.DWORD
	Status               ndr.DWORD `ndr:"retval"`
}

// RpcQuerySessionData calls RpcQuerySessionData (opnum 11) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcQuerySessionData(rpc ndr.Invoker, sessionId ndr.DWORD, type_ mststs.QUERY_SESSION_DATA_TYPE, pbInputData []uint8, cbInputData ndr.DWORD, cbSessionData ndr.DWORD) (PbSessionData []uint8, PcbReturnLength ndr.DWORD, PcbRequireBufferSize ndr.DWORD, err error) {
	req := &rpcQuerySessionDataRequest{
		SessionId:     sessionId,
		Type:          type_,
		PbInputData:   pbInputData,
		CbInputData:   cbInputData,
		CbSessionData: cbSessionData,
	}
	var resp rpcQuerySessionDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcQuerySessionData: %w", err)
		return
	}
	PbSessionData = resp.PbSessionData
	PcbReturnLength = resp.PcbReturnLength
	PcbRequireBufferSize = resp.PcbRequireBufferSize
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcQuerySessionData failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
