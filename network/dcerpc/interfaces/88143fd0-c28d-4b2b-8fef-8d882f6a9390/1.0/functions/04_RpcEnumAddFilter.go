package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcEnumAddFilterRequest carries the [in] parameters of RpcEnumAddFilter.
type rpcEnumAddFilterRequest struct {
	HEnum    mststs.ENUM_HANDLE
	HSubEnum mststs.ENUM_HANDLE
}

func (*rpcEnumAddFilterRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcEnumAddFilter }

// rpcEnumAddFilterResponse carries the [out] parameters and return value of RpcEnumAddFilter.
type rpcEnumAddFilterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcEnumAddFilter calls RpcEnumAddFilter (opnum 4) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcEnumAddFilter(rpc ndr.Invoker, hEnum mststs.ENUM_HANDLE, hSubEnum mststs.ENUM_HANDLE) (err error) {
	req := &rpcEnumAddFilterRequest{
		HEnum:    hEnum,
		HSubEnum: hSubEnum,
	}
	var resp rpcEnumAddFilterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumAddFilter: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcEnumAddFilter failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
