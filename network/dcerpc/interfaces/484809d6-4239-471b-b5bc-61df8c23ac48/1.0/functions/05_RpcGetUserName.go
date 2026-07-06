package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetUserNameRequest carries the [in] parameters of RpcGetUserName.
type rpcGetUserNameRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcGetUserNameRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetUserName }

// rpcGetUserNameResponse carries the [out] parameters and return value of RpcGetUserName.
type rpcGetUserNameResponse struct {
	PszUserName ndr.WSTR
	PszDomain   ndr.WSTR
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcGetUserName calls RpcGetUserName (opnum 5) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetUserName(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (PszUserName ndr.WSTR, PszDomain ndr.WSTR, err error) {
	req := &rpcGetUserNameRequest{
		HSession: hSession,
	}
	var resp rpcGetUserNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetUserName: %w", err)
		return
	}
	PszUserName = resp.PszUserName
	PszDomain = resp.PszDomain
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetUserName failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
