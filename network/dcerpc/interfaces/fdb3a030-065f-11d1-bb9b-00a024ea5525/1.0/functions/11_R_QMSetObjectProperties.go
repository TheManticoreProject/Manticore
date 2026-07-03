package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// r_QMSetObjectPropertiesRequest carries the [in] parameters of R_QMSetObjectProperties.
type r_QMSetObjectPropertiesRequest struct {
	PObjectFormat msmqmp.OBJECT_FORMAT
	Cp            ndr.DWORD
	AProp         []ndr.DWORD          `ndr:"unique,size_is=Cp"`
	ApVar         []msmqmq.PROPVARIANT `ndr:"unique,size_is=Cp"`
}

func (*r_QMSetObjectPropertiesRequest) Opnum() uint16 { return qmcomm.OpnumR_QMSetObjectProperties }

// r_QMSetObjectPropertiesResponse carries the [out] parameters and return value of R_QMSetObjectProperties.
type r_QMSetObjectPropertiesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMSetObjectProperties calls R_QMSetObjectProperties (opnum 11) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMSetObjectProperties(rpc ndr.Invoker, pObjectFormat msmqmp.OBJECT_FORMAT, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT) (err error) {
	req := &r_QMSetObjectPropertiesRequest{
		PObjectFormat: pObjectFormat,
		Cp:            cp,
		AProp:         aProp,
		ApVar:         apVar,
	}
	var resp r_QMSetObjectPropertiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMSetObjectProperties: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMSetObjectProperties failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
