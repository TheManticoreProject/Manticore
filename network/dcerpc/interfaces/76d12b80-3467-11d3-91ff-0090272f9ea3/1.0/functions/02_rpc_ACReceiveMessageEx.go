package functions

import (
	"fmt"

	qmcomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76d12b80-3467-11d3-91ff-0090272f9ea3/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// rpc_ACReceiveMessageExRequest carries the [in] parameters of rpc_ACReceiveMessageEx.
type rpc_ACReceiveMessageExRequest struct {
	HQMContext ndr.DWORD
	Ptb        msmqmp.CACTransferBufferV2
}

func (*rpc_ACReceiveMessageExRequest) Opnum() uint16 { return qmcomm2.Opnumrpc_ACReceiveMessageEx }

// rpc_ACReceiveMessageExResponse carries the [out] parameters and return value of rpc_ACReceiveMessageEx.
type rpc_ACReceiveMessageExResponse struct {
	Ptb    msmqmp.CACTransferBufferV2
	Status ndr.DWORD `ndr:"retval"`
}

// Rpc_ACReceiveMessageEx (opnum 2, [MS-MQMP] 3.1.5.4) is DEFERRED. It carries a
// CACTransferBufferV2, whose faithful NDR encoding requires pointer double-indirection the
// declarative codec does not yet support (see windows/protocols/ms-mqmp/CACTransferBufferV1.go
// and issue #801). The opnum and request/response shapes are present so the interface is
// accounted for; the call returns a not-implemented error until the codec gains
// double-indirection support and the buffer layout is validated against a live MSMQ server.
func Rpc_ACReceiveMessageEx(rpc ndr.Invoker, hQMContext ndr.DWORD, ptb msmqmp.CACTransferBufferV2) (Ptb msmqmp.CACTransferBufferV2, err error) {
	return msmqmp.CACTransferBufferV2{}, fmt.Errorf("rpc_ACReceiveMessageEx: not implemented — CACTransferBufferV2 requires NDR pointer double-indirection (issue #801)")
}
