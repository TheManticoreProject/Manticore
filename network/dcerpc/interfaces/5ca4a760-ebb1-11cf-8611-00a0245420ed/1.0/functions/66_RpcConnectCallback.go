package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcConnectCallbackRequest carries the [in] parameters of RpcConnectCallback.
type rpcConnectCallbackRequest struct {
	HServer     mststs.SERVER_HANDLE
	TimeOut     ndr.DWORD
	AddressType ndr.DWORD
	PAddress    []uint8 `ndr:"ref,size_is=AddressSize"`
	AddressSize ndr.DWORD
}

func (*rpcConnectCallbackRequest) Opnum() uint16 { return IcaApi.OpnumRpcConnectCallback }

// rpcConnectCallbackResponse carries the [out] parameters and return value of RpcConnectCallback.
type rpcConnectCallbackResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcConnectCallback calls RpcConnectCallback (opnum 66) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcConnectCallback(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, timeOut ndr.DWORD, addressType ndr.DWORD, pAddress []uint8, addressSize ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcConnectCallbackRequest{
		HServer:     hServer,
		TimeOut:     timeOut,
		AddressType: addressType,
		PAddress:    pAddress,
		AddressSize: addressSize,
	}
	var resp rpcConnectCallbackResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcConnectCallback: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcConnectCallback failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
