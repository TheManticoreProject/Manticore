package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// opnum215ReservedRequest carries the [in] parameters of Opnum215Reserved.
type opnum215ReservedRequest struct {
}

func (*opnum215ReservedRequest) Opnum() uint16 { return clusapi.Opnum215Reserved }

// opnum215ReservedResponse carries the [out] parameters and return value of Opnum215Reserved.
type opnum215ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum215Reserved calls Opnum215Reserved (opnum 215) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum215Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum215ReservedRequest{}
	var resp opnum215ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum215Reserved: %w", err)
		return
	}
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("Opnum215Reserved failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
