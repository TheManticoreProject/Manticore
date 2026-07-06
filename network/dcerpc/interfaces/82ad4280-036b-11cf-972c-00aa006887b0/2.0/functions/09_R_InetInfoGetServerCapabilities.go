package functions

// IDL source: [MS-IRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-irp/ed7e5940-9700-4a1f-8555-de29f99fe115
// A fetched copy is kept at ms-irp.idl in the interface directory.

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// r_InetInfoGetServerCapabilitiesRequest carries the [in] parameters of R_InetInfoGetServerCapabilities.
type r_InetInfoGetServerCapabilitiesRequest struct {
	PszServer  *ndr.WSTR `ndr:"unique"`
	DwReserved ndr.DWORD
}

func (*r_InetInfoGetServerCapabilitiesRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoGetServerCapabilities
}

// r_InetInfoGetServerCapabilitiesResponse carries the [out] parameters and return value of R_InetInfoGetServerCapabilities.
type r_InetInfoGetServerCapabilitiesResponse struct {
	PpCap  *msirp.INET_INFO_CAPABILITIES_STRUCT `ndr:"unique"`
	Status ndr.DWORD                            `ndr:"retval"`
}

// R_InetInfoGetServerCapabilities calls R_InetInfoGetServerCapabilities (opnum 9) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoGetServerCapabilities(rpc ndr.Invoker, pszServer *ndr.WSTR, dwReserved ndr.DWORD) (PpCap *msirp.INET_INFO_CAPABILITIES_STRUCT, err error) {
	req := &r_InetInfoGetServerCapabilitiesRequest{
		PszServer:  pszServer,
		DwReserved: dwReserved,
	}
	var resp r_InetInfoGetServerCapabilitiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoGetServerCapabilities: %w", err)
		return
	}
	PpCap = resp.PpCap
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoGetServerCapabilities failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
