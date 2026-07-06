package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	RCMListener "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/497d95a6-2d27-4bf5-9bbd-a6046957133c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcOpenListenerRequest carries the [in] parameters of RpcOpenListener.
type rpcOpenListenerRequest struct {
	SzListenerName ndr.WSTR
}

func (*rpcOpenListenerRequest) Opnum() uint16 { return RCMListener.OpnumRpcOpenListener }

// rpcOpenListenerResponse carries the [out] parameters and return value of RpcOpenListener.
type rpcOpenListenerResponse struct {
	PhListener mststs.HLISTENER
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcOpenListener calls RpcOpenListener (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcOpenListener(rpc ndr.Invoker, szListenerName ndr.WSTR) (PhListener mststs.HLISTENER, err error) {
	req := &rpcOpenListenerRequest{
		SzListenerName: szListenerName,
	}
	var resp rpcOpenListenerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenListener: %w", err)
		return
	}
	PhListener = resp.PhListener
	if uint32(resp.Status) != RCMListener.StatusSuccess {
		err = fmt.Errorf("RpcOpenListener failed: %s", RCMListener.StatusString(uint32(resp.Status)))
	}
	return
}
