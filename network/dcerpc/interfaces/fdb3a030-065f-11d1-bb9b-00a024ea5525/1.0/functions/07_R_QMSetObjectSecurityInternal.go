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

// r_QMSetObjectSecurityInternalRequest carries the [in] parameters of R_QMSetObjectSecurityInternal.
type r_QMSetObjectSecurityInternalRequest struct {
	PObjectFormat       msmqmp.OBJECT_FORMAT
	SecurityInformation ndr.DWORD
	SDSize              ndr.DWORD
	PSecurityDescriptor []uint8 `ndr:"unique,size_is=SDSize"`
}

func (*r_QMSetObjectSecurityInternalRequest) Opnum() uint16 {
	return qmcomm.OpnumR_QMSetObjectSecurityInternal
}

// r_QMSetObjectSecurityInternalResponse carries the [out] parameters and return value of R_QMSetObjectSecurityInternal.
type r_QMSetObjectSecurityInternalResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMSetObjectSecurityInternal calls R_QMSetObjectSecurityInternal (opnum 7) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMSetObjectSecurityInternal(rpc ndr.Invoker, pObjectFormat msmqmp.OBJECT_FORMAT, securityInformation ndr.DWORD, sDSize ndr.DWORD, pSecurityDescriptor []uint8) (err error) {
	req := &r_QMSetObjectSecurityInternalRequest{
		PObjectFormat:       pObjectFormat,
		SecurityInformation: securityInformation,
		SDSize:              sDSize,
		PSecurityDescriptor: pSecurityDescriptor,
	}
	var resp r_QMSetObjectSecurityInternalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMSetObjectSecurityInternal: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMSetObjectSecurityInternal failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
