package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmp"
)

// r_QMObjectPathToObjectFormatRequest carries the [in] parameters of R_QMObjectPathToObjectFormat.
type r_QMObjectPathToObjectFormatRequest struct {
	LpwcsPathName ndr.WSTR
	PObjectFormat msmqmp.OBJECT_FORMAT
}

func (*r_QMObjectPathToObjectFormatRequest) Opnum() uint16 {
	return qmcomm.OpnumR_QMObjectPathToObjectFormat
}

// r_QMObjectPathToObjectFormatResponse carries the [out] parameters and return value of R_QMObjectPathToObjectFormat.
type r_QMObjectPathToObjectFormatResponse struct {
	PObjectFormat msmqmp.OBJECT_FORMAT
	Status        ndr.DWORD `ndr:"retval"`
}

// R_QMObjectPathToObjectFormat calls R_QMObjectPathToObjectFormat (opnum 12) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMObjectPathToObjectFormat(rpc ndr.Invoker, lpwcsPathName ndr.WSTR, pObjectFormat msmqmp.OBJECT_FORMAT) (PObjectFormat msmqmp.OBJECT_FORMAT, err error) {
	req := &r_QMObjectPathToObjectFormatRequest{
		LpwcsPathName: lpwcsPathName,
		PObjectFormat: pObjectFormat,
	}
	var resp r_QMObjectPathToObjectFormatResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMObjectPathToObjectFormat: %w", err)
		return
	}
	PObjectFormat = resp.PObjectFormat
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMObjectPathToObjectFormat failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
