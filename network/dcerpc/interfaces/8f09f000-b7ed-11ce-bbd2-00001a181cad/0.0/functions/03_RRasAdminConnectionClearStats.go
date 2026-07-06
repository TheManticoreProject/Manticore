package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRasAdminConnectionClearStatsRequest carries the [in] parameters of RRasAdminConnectionClearStats.
type rRasAdminConnectionClearStatsRequest struct {
	HDimConnection ndr.DWORD
}

func (*rRasAdminConnectionClearStatsRequest) Opnum() uint16 {
	return dimsvc.OpnumRRasAdminConnectionClearStats
}

// rRasAdminConnectionClearStatsResponse carries the [out] parameters and return value of RRasAdminConnectionClearStats.
type rRasAdminConnectionClearStatsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminConnectionClearStats calls RRasAdminConnectionClearStats (opnum 3) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionClearStats(rpc ndr.Invoker, hDimConnection ndr.DWORD) (err error) {
	req := &rRasAdminConnectionClearStatsRequest{
		HDimConnection: hDimConnection,
	}
	var resp rRasAdminConnectionClearStatsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionClearStats: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionClearStats failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
