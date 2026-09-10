package functions

// IDL source: [MS-RAIW] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-raiw/e59461f5-5486-4ec3-9ad6-14b784c1ecd6
// A fetched copy is kept at ms-raiw.idl in the interface directory.

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsGetNameAndAddRequest carries the [in] parameters of R_WinsGetNameAndAdd.
type r_WinsGetNameAndAddRequest struct {
}

func (*r_WinsGetNameAndAddRequest) Opnum() uint16 { return winsif.OpnumR_WinsGetNameAndAdd }

// r_WinsGetNameAndAddResponse carries the [out] parameters and return value of R_WinsGetNameAndAdd.
type r_WinsGetNameAndAddResponse struct {
	PWinsAdd msraiw.WINSINTF_ADD_T
	PUncName ndr.STR
	Status   ndr.DWORD `ndr:"retval"`
}

// R_WinsGetNameAndAdd calls R_WinsGetNameAndAdd (opnum 13) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsGetNameAndAdd(rpc ndr.Invoker) (PWinsAdd msraiw.WINSINTF_ADD_T, PUncName ndr.STR, err error) {
	req := &r_WinsGetNameAndAddRequest{}
	var resp r_WinsGetNameAndAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsGetNameAndAdd: %w", err)
		return
	}
	PWinsAdd = resp.PWinsAdd
	PUncName = resp.PUncName
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("R_WinsGetNameAndAdd failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
