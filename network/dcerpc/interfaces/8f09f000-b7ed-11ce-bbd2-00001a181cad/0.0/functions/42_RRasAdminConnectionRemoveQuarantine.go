package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRasAdminConnectionRemoveQuarantineRequest carries the [in] parameters of RRasAdminConnectionRemoveQuarantine.
type rRasAdminConnectionRemoveQuarantineRequest struct {
	HRasConnection ndr.DWORD
	FIsIpAddress   ndr.BOOL
}

func (*rRasAdminConnectionRemoveQuarantineRequest) Opnum() uint16 {
	return dimsvc.OpnumRRasAdminConnectionRemoveQuarantine
}

// rRasAdminConnectionRemoveQuarantineResponse carries the [out] parameters and return value of RRasAdminConnectionRemoveQuarantine.
type rRasAdminConnectionRemoveQuarantineResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminConnectionRemoveQuarantine calls RRasAdminConnectionRemoveQuarantine (opnum 42) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionRemoveQuarantine(rpc ndr.Invoker, hRasConnection ndr.DWORD, fIsIpAddress ndr.BOOL) (err error) {
	req := &rRasAdminConnectionRemoveQuarantineRequest{
		HRasConnection: hRasConnection,
		FIsIpAddress:   fIsIpAddress,
	}
	var resp rRasAdminConnectionRemoveQuarantineResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionRemoveQuarantine: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionRemoveQuarantine failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
