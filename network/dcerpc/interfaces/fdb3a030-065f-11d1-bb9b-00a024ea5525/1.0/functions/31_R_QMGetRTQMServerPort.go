package functions

import (
	"fmt"

	qmcomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/fdb3a030-065f-11d1-bb9b-00a024ea5525/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_QMGetRTQMServerPortRequest carries the [in] parameters of R_QMGetRTQMServerPort.
type r_QMGetRTQMServerPortRequest struct {
	FIP ndr.DWORD
}

func (*r_QMGetRTQMServerPortRequest) Opnum() uint16 { return qmcomm.OpnumR_QMGetRTQMServerPort }

// r_QMGetRTQMServerPortResponse carries the return value of R_QMGetRTQMServerPort. Unlike
// the other qmcomm methods, the DWORD return is a port number, not an HRESULT.
type r_QMGetRTQMServerPortResponse struct {
	Port ndr.DWORD `ndr:"retval"`
}

// R_QMGetRTQMServerPort calls R_QMGetRTQMServerPort (opnum 31, [MS-MQMP] 3.1.4.24) and
// returns the port the RT-QM server listens on: the TCP/IP port when fIP is non-zero, or
// the IPX/SPX port when fIP is zero. The method returns the port directly as a DWORD
// rather than an HRESULT, so there is no status check.
func R_QMGetRTQMServerPort(rpc ndr.Invoker, fIP ndr.DWORD) (port ndr.DWORD, err error) {
	req := &r_QMGetRTQMServerPortRequest{
		FIP: fIP,
	}
	var resp r_QMGetRTQMServerPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("R_QMGetRTQMServerPort: %w", err)
	}
	return resp.Port, nil
}
