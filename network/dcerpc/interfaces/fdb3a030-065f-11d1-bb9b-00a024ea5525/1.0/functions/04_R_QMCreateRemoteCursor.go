package functions

// IDL source: [MS-MQMP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmp/a54c09de-1d72-47f0-9184-d7e5046b2ba1
// A fetched copy is kept at ms-mqmp.idl in the interface directory.

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// r_QMCreateRemoteCursorRequest carries the [in] parameters of R_QMCreateRemoteCursor.
type r_QMCreateRemoteCursorRequest struct {
	Ptb1   msmqmp.CACTransferBufferV1
	HQueue ndr.DWORD
}

func (*r_QMCreateRemoteCursorRequest) Opnum() uint16 { return qmcomm.OpnumR_QMCreateRemoteCursor }

// r_QMCreateRemoteCursorResponse carries the [out] parameters and return value of R_QMCreateRemoteCursor.
type r_QMCreateRemoteCursorResponse struct {
	PhCursor ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// R_QMCreateRemoteCursor (opnum 4, [MS-MQMP] 3.1.4.4) is DEFERRED. It carries a
// CACTransferBufferV1, whose faithful NDR encoding requires pointer double-indirection the
// declarative codec does not yet support (see windows/protocols/ms-mqmp/CACTransferBufferV1.go
// and issue #801). The opnum and request/response shapes are present so the interface is
// accounted for; the call returns a not-implemented error until the codec gains
// double-indirection support and the buffer layout is validated against a live MSMQ server.
func R_QMCreateRemoteCursor(rpc ndr.Invoker, ptb1 msmqmp.CACTransferBufferV1, hQueue ndr.DWORD) (PhCursor ndr.DWORD, err error) {
	return 0, fmt.Errorf("R_QMCreateRemoteCursor: not implemented — CACTransferBufferV1 requires NDR pointer double-indirection (issue #801)")
}
