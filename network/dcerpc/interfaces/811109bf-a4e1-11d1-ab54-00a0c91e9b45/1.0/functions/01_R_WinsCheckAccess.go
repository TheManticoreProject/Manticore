package functions

import (
	"fmt"

	winsi2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/811109bf-a4e1-11d1-ab54-00a0c91e9b45/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_WinsCheckAccessRequest carries the [in] parameters of R_WinsCheckAccess.
type r_WinsCheckAccessRequest struct {
}

func (*r_WinsCheckAccessRequest) Opnum() uint16 { return winsi2.OpnumR_WinsCheckAccess }

// r_WinsCheckAccessResponse carries the [out] parameters and return value of R_WinsCheckAccess.
type r_WinsCheckAccessResponse struct {
	Access ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsCheckAccess calls R_WinsCheckAccess (opnum 1) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsCheckAccess(rpc ndr.Invoker) (Access ndr.DWORD, err error) {
	req := &r_WinsCheckAccessRequest{}
	var resp r_WinsCheckAccessResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsCheckAccess: %w", err)
		return
	}
	Access = resp.Access
	if uint32(resp.Status) != winsi2.StatusSuccess {
		err = fmt.Errorf("R_WinsCheckAccess failed: %s", winsi2.StatusString(uint32(resp.Status)))
	}
	return
}
