package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetActivityIdRequest carries the [in] parameters of RpcGetActivityId.
type rpcGetActivityIdRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcGetActivityIdRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetActivityId }

// rpcGetActivityIdResponse carries the [out] parameters and return value of RpcGetActivityId.
type rpcGetActivityIdResponse struct {
	PActivityId *guid.GUID `ndr:"unique"`
	Status      ndr.DWORD  `ndr:"retval"`
}

// RpcGetActivityId calls RpcGetActivityId (opnum 21) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetActivityId(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (PActivityId *guid.GUID, err error) {
	req := &rpcGetActivityIdRequest{
		HSession: hSession,
	}
	var resp rpcGetActivityIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetActivityId: %w", err)
		return
	}
	PActivityId = resp.PActivityId
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetActivityId failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
