package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_WinsDoStaticInitRequest carries the [in] parameters of R_WinsDoStaticInit.
type r_WinsDoStaticInitRequest struct {
	PDataFilePath *ndr.WSTR `ndr:"unique"`
	FDel          ndr.DWORD
}

func (*r_WinsDoStaticInitRequest) Opnum() uint16 { return winsif.OpnumR_WinsDoStaticInit }

// r_WinsDoStaticInitResponse carries the [out] parameters and return value of R_WinsDoStaticInit.
type r_WinsDoStaticInitResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsDoStaticInit calls R_WinsDoStaticInit (opnum 3) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsDoStaticInit(rpc ndr.Invoker, pDataFilePath *ndr.WSTR, fDel ndr.DWORD) (err error) {
	req := &r_WinsDoStaticInitRequest{
		PDataFilePath: pDataFilePath,
		FDel:          fDel,
	}
	var resp r_WinsDoStaticInitResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsDoStaticInit: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsDoStaticInit failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
