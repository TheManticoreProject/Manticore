package functions

// IDL source: [MS-IRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-irp/ed7e5940-9700-4a1f-8555-de29f99fe115
// A fetched copy is kept at ms-irp.idl in the interface directory.

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_W3ClearStatistics2Request carries the [in] parameters of R_W3ClearStatistics2.
type r_W3ClearStatistics2Request struct {
	PszServer  *ndr.WSTR `ndr:"unique"`
	DwInstance ndr.DWORD
}

func (*r_W3ClearStatistics2Request) Opnum() uint16 { return inetinfo.OpnumR_W3ClearStatistics2 }

// r_W3ClearStatistics2Response carries the [out] parameters and return value of R_W3ClearStatistics2.
type r_W3ClearStatistics2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_W3ClearStatistics2 calls R_W3ClearStatistics2 (opnum 11) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_W3ClearStatistics2(rpc ndr.Invoker, pszServer *ndr.WSTR, dwInstance ndr.DWORD) (err error) {
	req := &r_W3ClearStatistics2Request{
		PszServer:  pszServer,
		DwInstance: dwInstance,
	}
	var resp r_W3ClearStatistics2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_W3ClearStatistics2: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_W3ClearStatistics2 failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
