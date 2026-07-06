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

// r_InetInfoGetAdminInformationRequest carries the [in] parameters of R_InetInfoGetAdminInformation.
type r_InetInfoGetAdminInformationRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
}

func (*r_InetInfoGetAdminInformationRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoGetAdminInformation
}

// r_InetInfoGetAdminInformationResponse carries the [out] parameters and return value of R_InetInfoGetAdminInformation.
type r_InetInfoGetAdminInformationResponse struct {
	PpConfig *msirp.INET_INFO_CONFIG_INFO `ndr:"unique"`
	Status   ndr.DWORD                    `ndr:"retval"`
}

// R_InetInfoGetAdminInformation calls R_InetInfoGetAdminInformation (opnum 1) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoGetAdminInformation(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD) (PpConfig *msirp.INET_INFO_CONFIG_INFO, err error) {
	req := &r_InetInfoGetAdminInformationRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
	}
	var resp r_InetInfoGetAdminInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoGetAdminInformation: %w", err)
		return
	}
	PpConfig = resp.PpConfig
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoGetAdminInformation failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
