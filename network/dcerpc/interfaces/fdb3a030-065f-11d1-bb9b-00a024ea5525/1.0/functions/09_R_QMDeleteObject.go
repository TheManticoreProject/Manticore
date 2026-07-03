package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// r_QMDeleteObjectRequest carries the [in] parameters of R_QMDeleteObject.
type r_QMDeleteObjectRequest struct {
	PObjectFormat msmqmp.OBJECT_FORMAT
}

func (*r_QMDeleteObjectRequest) Opnum() uint16 { return qmcomm.OpnumR_QMDeleteObject }

// r_QMDeleteObjectResponse carries the [out] parameters and return value of R_QMDeleteObject.
type r_QMDeleteObjectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_QMDeleteObject calls R_QMDeleteObject (opnum 9) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMDeleteObject(rpc ndr.Invoker, pObjectFormat msmqmp.OBJECT_FORMAT) (err error) {
	req := &r_QMDeleteObjectRequest{
		PObjectFormat: pObjectFormat,
	}
	var resp r_QMDeleteObjectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMDeleteObject: %w", err)
		return
	}
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMDeleteObject failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
