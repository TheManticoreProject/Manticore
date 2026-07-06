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

// r_InetInfoGetVersionRequest carries the [in] parameters of R_InetInfoGetVersion.
type r_InetInfoGetVersionRequest struct {
	PszServer  *ndr.WSTR `ndr:"unique"`
	DwReserved ndr.DWORD
}

func (*r_InetInfoGetVersionRequest) Opnum() uint16 { return inetinfo.OpnumR_InetInfoGetVersion }

// r_InetInfoGetVersionResponse carries the [out] parameters and return value of R_InetInfoGetVersion.
type r_InetInfoGetVersionResponse struct {
	PdwVersion ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// R_InetInfoGetVersion calls R_InetInfoGetVersion (opnum 0) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoGetVersion(rpc ndr.Invoker, pszServer *ndr.WSTR, dwReserved ndr.DWORD) (PdwVersion ndr.DWORD, err error) {
	req := &r_InetInfoGetVersionRequest{
		PszServer:  pszServer,
		DwReserved: dwReserved,
	}
	var resp r_InetInfoGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoGetVersion: %w", err)
		return
	}
	PdwVersion = resp.PdwVersion
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoGetVersion failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
