package functions

// IDL source: [MS-MQMQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqmq/56cc73e0-f57a-4bd9-aa45-861be5b85049
// A fetched copy is kept at ms-mqmq.idl in the interface directory.

import (
	"fmt"

	qmmgmt "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/41208ee0-e970-11d1-9b9e-00e02c064c39/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
	msmqmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmr"
)

// r_QMMgmtGetInfoRequest carries the [in]/[in,out] parameters of R_QMMgmtGetInfo
// ([MS-MQMR] 3.1.4.1). pObjectFormat is a top-level [ref] pointer, so it is transmitted
// inline. aProp and apVar are [ref] pointers to conformant arrays whose element count is
// cp (1..128); apVar is [in,out], so it appears in both request and response.
type r_QMMgmtGetInfoRequest struct {
	PObjectFormat msmqmr.MGMT_OBJECT
	Cp            ndr.DWORD
	AProp         []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar         []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
}

func (*r_QMMgmtGetInfoRequest) Opnum() uint16 { return qmmgmt.OpnumR_QMMgmtGetInfo }

// r_QMMgmtGetInfoResponse carries the [in,out] apVar array and the HRESULT return value.
type r_QMMgmtGetInfoResponse struct {
	ApVar  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	Status ndr.DWORD            `ndr:"retval"`
}

// R_QMMgmtGetInfo requests information on an MSMQ installation or a specific queue
// (opnum 0) ([MS-MQMR] 3.1.4.1). pObjectFormat selects the target (MGMT_MACHINE for the
// installation, MGMT_QUEUE for a queue). aProp holds cp property identifiers ([MS-MQMR]
// 2.2.3); apVar must hold cp elements each set to VT_NULL on input and receives the
// property values on output. The filled apVar array is returned.
func R_QMMgmtGetInfo(rpc ndr.Invoker, pObjectFormat msmqmr.MGMT_OBJECT, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT) (ApVar []msmqmq.PROPVARIANT, err error) {
	req := &r_QMMgmtGetInfoRequest{
		PObjectFormat: pObjectFormat,
		Cp:            cp,
		AProp:         aProp,
		ApVar:         apVar,
	}
	var resp r_QMMgmtGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMMgmtGetInfo: %w", err)
		return
	}
	ApVar = resp.ApVar
	if uint32(resp.Status) != qmmgmt.StatusSuccess {
		err = fmt.Errorf("R_QMMgmtGetInfo failed: %s", qmmgmt.StatusString(uint32(resp.Status)))
	}
	return
}
