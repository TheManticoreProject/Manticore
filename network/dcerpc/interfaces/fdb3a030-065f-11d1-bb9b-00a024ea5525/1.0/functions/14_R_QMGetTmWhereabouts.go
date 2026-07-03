package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_QMGetTmWhereaboutsRequest carries the [in] parameters of R_QMGetTmWhereabouts.
type r_QMGetTmWhereaboutsRequest struct {
	CbBufSize ndr.DWORD
}

func (*r_QMGetTmWhereaboutsRequest) Opnum() uint16 { return qmcomm.OpnumR_QMGetTmWhereabouts }

// r_QMGetTmWhereaboutsResponse carries the [out] parameters and return value of R_QMGetTmWhereabouts.
type r_QMGetTmWhereaboutsResponse struct {
	PbWhereabouts  []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbWhereabouts ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// R_QMGetTmWhereabouts calls R_QMGetTmWhereabouts (opnum 14) ([MS-MQMP] — verify the parameter
// modeling and status handling).
func R_QMGetTmWhereabouts(rpc ndr.Invoker, cbBufSize ndr.DWORD) (PbWhereabouts []uint8, PcbWhereabouts ndr.DWORD, err error) {
	req := &r_QMGetTmWhereaboutsRequest{
		CbBufSize: cbBufSize,
	}
	var resp r_QMGetTmWhereaboutsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_QMGetTmWhereabouts: %w", err)
		return
	}
	PbWhereabouts = resp.PbWhereabouts
	PcbWhereabouts = resp.PcbWhereabouts
	if uint32(resp.Status) != qmcomm.StatusSuccess {
		err = fmt.Errorf("R_QMGetTmWhereabouts failed: %s", qmcomm.StatusString(uint32(resp.Status)))
	}
	return
}
