package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetSessionIdsRequest carries the [in] parameters of RpcGetSessionIds.
type rpcGetSessionIdsRequest struct {
	Filter     mststs.SESSION_FILTER
	MaxEntries ndr.DWORD
}

func (*rpcGetSessionIdsRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcGetSessionIds }

// rpcGetSessionIdsResponse carries the [out] parameters and return value of RpcGetSessionIds.
type rpcGetSessionIdsResponse struct {
	PSessionIds  []int32 `ndr:"unique,conformant"`
	PcSessionIds ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcGetSessionIds calls RpcGetSessionIds (opnum 8) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionIds(rpc ndr.Invoker, filter mststs.SESSION_FILTER, maxEntries ndr.DWORD) (PSessionIds []int32, PcSessionIds ndr.DWORD, err error) {
	req := &rpcGetSessionIdsRequest{
		Filter:     filter,
		MaxEntries: maxEntries,
	}
	var resp rpcGetSessionIdsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionIds: %w", err)
		return
	}
	PSessionIds = resp.PSessionIds
	PcSessionIds = resp.PcSessionIds
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionIds failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
