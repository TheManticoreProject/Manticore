package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_WinsDoScavengingRequest carries the [in] parameters of R_WinsDoScavenging.
type r_WinsDoScavengingRequest struct {
}

func (*r_WinsDoScavengingRequest) Opnum() uint16 { return winsif.OpnumR_WinsDoScavenging }

// r_WinsDoScavengingResponse carries the [out] parameters and return value of R_WinsDoScavenging.
type r_WinsDoScavengingResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsDoScavenging calls R_WinsDoScavenging (opnum 4) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsDoScavenging(rpc ndr.Invoker) (err error) {
	req := &r_WinsDoScavengingRequest{}
	var resp r_WinsDoScavengingResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsDoScavenging: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsDoScavenging failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
