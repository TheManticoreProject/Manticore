package functions

// IDL source: [MS-MQMQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmq/56cc73e0-f57a-4bd9-aa45-861be5b85049
// A fetched copy is kept at ms-mqmq.idl in the interface directory.

import (
	"fmt"

	qmmgmt "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/41208ee0-e970-11d1-9b9e-00e02c064c39/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmr"
)

// r_QMMgmtActionRequest carries the [in] parameters of R_QMMgmtAction ([MS-MQMR]
// 3.1.4.2). pObjectFormat is a top-level [ref] pointer transmitted inline; lpwszAction is
// a [ref] pointer to a null-terminated Unicode string naming the action to perform.
type r_QMMgmtActionRequest struct {
	PObjectFormat msmqmr.MGMT_OBJECT
	LpwszAction   ndr.WSTR
}

func (*r_QMMgmtActionRequest) Opnum() uint16 { return qmmgmt.OpnumR_QMMgmtAction }

// r_QMMgmtActionResponse carries the HRESULT return value of R_QMMgmtAction.
type r_QMMgmtActionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMMgmtAction requests the server to perform a management function on a queue or MSMQ
// installation (opnum 1) ([MS-MQMR] 3.1.4.2). pObjectFormat selects the target and action
// is one of the documented, case-insensitive verbs (see the qmmgmt.Action* constants,
// e.g. "PAUSE"/"RESUME" for outgoing queues, "CONNECT"/"DISCONNECT"/"TIDY" for the
// machine).
func R_QMMgmtAction(rpc ndr.Invoker, pObjectFormat msmqmr.MGMT_OBJECT, action string) error {
	req := &r_QMMgmtActionRequest{
		PObjectFormat: pObjectFormat,
		LpwszAction:   ndr.WSTR(action),
	}
	var resp r_QMMgmtActionResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("R_QMMgmtAction: %w", err)
	}
	if uint32(resp.Status) != qmmgmt.StatusSuccess {
		return fmt.Errorf("R_QMMgmtAction failed: %s", qmmgmt.StatusString(uint32(resp.Status)))
	}
	return nil
}
