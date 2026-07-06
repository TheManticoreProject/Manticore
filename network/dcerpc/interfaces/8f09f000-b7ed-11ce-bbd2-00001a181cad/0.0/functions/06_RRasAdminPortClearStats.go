package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRasAdminPortClearStatsRequest carries the [in] parameters of RRasAdminPortClearStats.
type rRasAdminPortClearStatsRequest struct {
	HPort ndr.DWORD
}

func (*rRasAdminPortClearStatsRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminPortClearStats }

// rRasAdminPortClearStatsResponse carries the [out] parameters and return value of RRasAdminPortClearStats.
type rRasAdminPortClearStatsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminPortClearStats calls RRasAdminPortClearStats (opnum 6) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminPortClearStats(rpc ndr.Invoker, hPort ndr.DWORD) (err error) {
	req := &rRasAdminPortClearStatsRequest{
		HPort: hPort,
	}
	var resp rRasAdminPortClearStatsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminPortClearStats: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminPortClearStats failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
