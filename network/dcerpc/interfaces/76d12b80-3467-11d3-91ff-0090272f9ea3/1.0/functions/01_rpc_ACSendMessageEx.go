package functions

// IDL source: [MS-MQMP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmp/a54c09de-1d72-47f0-9184-d7e5046b2ba1
// A fetched copy is kept at ms-mqmp.idl in the interface directory.

import (
	"fmt"

	qmcomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76d12b80-3467-11d3-91ff-0090272f9ea3/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// rpc_ACSendMessageExRequest carries the [in] parameters of rpc_ACSendMessageEx.
type rpc_ACSendMessageExRequest struct {
	HQueue     msmqmp.RPC_QUEUE_HANDLE
	Ptb        msmqmp.CACTransferBufferV2
	PMessageID *msmqmq.OBJECTID `ndr:"unique"`
}

func (*rpc_ACSendMessageExRequest) Opnum() uint16 { return qmcomm2.Opnumrpc_ACSendMessageEx }

// rpc_ACSendMessageExResponse carries the [out] parameters and return value of rpc_ACSendMessageEx.
type rpc_ACSendMessageExResponse struct {
	PMessageID *msmqmq.OBJECTID `ndr:"unique"`
	Status     ndr.DWORD        `ndr:"retval"`
}

// Rpc_ACSendMessageEx (opnum 1, [MS-MQMP] 3.1.5.3) is DEFERRED. It carries a
// CACTransferBufferV2, whose faithful NDR encoding requires pointer double-indirection the
// declarative codec does not yet support (see windows/protocols/ms-mqmp/CACTransferBufferV1.go
// and issue #801). The opnum and request/response shapes are present so the interface is
// accounted for; the call returns a not-implemented error until the codec gains
// double-indirection support and the buffer layout is validated against a live MSMQ server.
func Rpc_ACSendMessageEx(rpc ndr.Invoker, hQueue msmqmp.RPC_QUEUE_HANDLE, ptb msmqmp.CACTransferBufferV2, pMessageID *msmqmq.OBJECTID) (PMessageID *msmqmq.OBJECTID, err error) {
	return nil, fmt.Errorf("rpc_ACSendMessageEx: not implemented — CACTransferBufferV2 requires NDR pointer double-indirection (issue #801)")
}
