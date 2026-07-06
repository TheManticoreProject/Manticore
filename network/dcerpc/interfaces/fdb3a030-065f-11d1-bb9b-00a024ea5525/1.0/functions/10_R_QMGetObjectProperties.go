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
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMGetObjectPropertiesRequest carries the [in] parameters of R_QMGetObjectProperties.
type r_QMGetObjectPropertiesRequest struct {
	PObjectFormat msmqmp.OBJECT_FORMAT
	Cp            ndr.DWORD
	AProp         []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar         []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
}

func (*r_QMGetObjectPropertiesRequest) Opnum() uint16 { return qmcomm.OpnumR_QMGetObjectProperties }

// r_QMGetObjectPropertiesResponse carries the [out] parameters and return value of R_QMGetObjectProperties.
type r_QMGetObjectPropertiesResponse struct {
	ApVar  []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	Status ndr.DWORD            `ndr:"retval"`
}

// R_QMGetObjectProperties calls R_QMGetObjectProperties (opnum 10) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMGetObjectProperties(rpc ndr.Invoker, pObjectFormat msmqmp.OBJECT_FORMAT, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT) (ApVar []msmqmq.PROPVARIANT, err error) {
	req := &r_QMGetObjectPropertiesRequest{
		PObjectFormat: pObjectFormat,
		Cp:            cp,
		AProp:         aProp,
		ApVar:         apVar,
	}
	var resp r_QMGetObjectPropertiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMGetObjectProperties: %w", err)
		return
	}
	ApVar = resp.ApVar
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMGetObjectProperties failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
