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

// r_QMGetObjectSecurityInternalRequest carries the [in] parameters of R_QMGetObjectSecurityInternal.
type r_QMGetObjectSecurityInternalRequest struct {
	PObjectFormat        msmqmp.OBJECT_FORMAT
	RequestedInformation ndr.DWORD
	NLength              ndr.DWORD
}

func (*r_QMGetObjectSecurityInternalRequest) Opnum() uint16 {
	return qmcomm.OpnumR_QMGetObjectSecurityInternal
}

// r_QMGetObjectSecurityInternalResponse carries the [out] parameters and return value of R_QMGetObjectSecurityInternal.
type r_QMGetObjectSecurityInternalResponse struct {
	PSecurityDescriptor []uint8 `ndr:"ref,size_is=NLength"`
	LpnLengthNeeded     ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// R_QMGetObjectSecurityInternal calls R_QMGetObjectSecurityInternal (opnum 8) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMGetObjectSecurityInternal(rpc ndr.Invoker, pObjectFormat msmqmp.OBJECT_FORMAT, requestedInformation ndr.DWORD, nLength ndr.DWORD) (PSecurityDescriptor []uint8, LpnLengthNeeded ndr.DWORD, err error) {
	req := &r_QMGetObjectSecurityInternalRequest{
		PObjectFormat:        pObjectFormat,
		RequestedInformation: requestedInformation,
		NLength:              nLength,
	}
	var resp r_QMGetObjectSecurityInternalResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMGetObjectSecurityInternal: %w", err)
		return
	}
	PSecurityDescriptor = resp.PSecurityDescriptor
	LpnLengthNeeded = resp.LpnLengthNeeded
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMGetObjectSecurityInternal failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
