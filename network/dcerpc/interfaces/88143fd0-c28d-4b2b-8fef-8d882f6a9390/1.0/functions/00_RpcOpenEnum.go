package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcOpenEnumRequest carries the [in] parameters of RpcOpenEnum.
type rpcOpenEnumRequest struct {
}

func (*rpcOpenEnumRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcOpenEnum }

// rpcOpenEnumResponse carries the [out] parameters and return value of RpcOpenEnum.
type rpcOpenEnumResponse struct {
	PhEnum mststs.ENUM_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

// RpcOpenEnum calls RpcOpenEnum (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcOpenEnum(rpc ndr.Invoker) (PhEnum mststs.ENUM_HANDLE, err error) {
	req := &rpcOpenEnumRequest{}
	var resp rpcOpenEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenEnum: %w", err)
		return
	}
	PhEnum = resp.PhEnum
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcOpenEnum failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
