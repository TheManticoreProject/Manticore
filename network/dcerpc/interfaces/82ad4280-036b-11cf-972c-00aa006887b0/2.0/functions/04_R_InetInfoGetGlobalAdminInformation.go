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

// r_InetInfoGetGlobalAdminInformationRequest carries the [in] parameters of R_InetInfoGetGlobalAdminInformation.
type r_InetInfoGetGlobalAdminInformationRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
}

func (*r_InetInfoGetGlobalAdminInformationRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoGetGlobalAdminInformation
}

// r_InetInfoGetGlobalAdminInformationResponse carries the [out] parameters and return value of R_InetInfoGetGlobalAdminInformation.
type r_InetInfoGetGlobalAdminInformationResponse struct {
	PpConfig *msirp.INET_INFO_GLOBAL_CONFIG_INFO `ndr:"unique"`
	Status   ndr.DWORD                           `ndr:"retval"`
}

// R_InetInfoGetGlobalAdminInformation calls R_InetInfoGetGlobalAdminInformation (opnum 4) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoGetGlobalAdminInformation(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD) (PpConfig *msirp.INET_INFO_GLOBAL_CONFIG_INFO, err error) {
	req := &r_InetInfoGetGlobalAdminInformationRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
	}
	var resp r_InetInfoGetGlobalAdminInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoGetGlobalAdminInformation: %w", err)
		return
	}
	PpConfig = resp.PpConfig
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoGetGlobalAdminInformation failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
