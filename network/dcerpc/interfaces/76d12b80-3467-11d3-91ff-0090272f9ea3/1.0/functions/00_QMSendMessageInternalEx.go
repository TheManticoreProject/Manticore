package functions

import (
	"fmt"

	qmcomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76d12b80-3467-11d3-91ff-0090272f9ea3/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// qMSendMessageInternalExRequest carries the [in] parameters of QMSendMessageInternalEx.
type qMSendMessageInternalExRequest struct {
	PQueueFormat msmqmq.QUEUE_FORMAT
	Ptb          msmqmp.CACTransferBufferV2
	PMessageID   *msmqmq.OBJECTID `ndr:"unique"`
}

func (*qMSendMessageInternalExRequest) Opnum() uint16 { return qmcomm2.OpnumQMSendMessageInternalEx }

// qMSendMessageInternalExResponse carries the [out] parameters and return value of QMSendMessageInternalEx.
type qMSendMessageInternalExResponse struct {
	PMessageID *msmqmq.OBJECTID `ndr:"unique"`
	Status     ndr.DWORD        `ndr:"retval"`
}

// QMSendMessageInternalEx (opnum 0, [MS-MQMP] 3.1.5.2) is DEFERRED. It carries a
// CACTransferBufferV2, whose faithful NDR encoding requires pointer double-indirection the
// declarative codec does not yet support (see windows/protocols/ms-mqmp/CACTransferBufferV1.go
// and issue #801). The opnum and request/response shapes are present so the interface is
// accounted for; the call returns a not-implemented error until the codec gains
// double-indirection support and the buffer layout is validated against a live MSMQ server.
func QMSendMessageInternalEx(rpc ndr.Invoker, pQueueFormat msmqmq.QUEUE_FORMAT, ptb msmqmp.CACTransferBufferV2, pMessageID *msmqmq.OBJECTID) (PMessageID *msmqmq.OBJECTID, err error) {
	return nil, fmt.Errorf("QMSendMessageInternalEx: not implemented — CACTransferBufferV2 requires NDR pointer double-indirection (issue #801)")
}
