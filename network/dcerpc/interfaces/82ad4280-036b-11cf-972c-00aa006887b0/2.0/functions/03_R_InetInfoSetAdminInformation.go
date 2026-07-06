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

// r_InetInfoSetAdminInformationRequest carries the [in] parameters of R_InetInfoSetAdminInformation.
type r_InetInfoSetAdminInformationRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
	PConfig      msirp.INET_INFO_CONFIG_INFO
}

func (*r_InetInfoSetAdminInformationRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoSetAdminInformation
}

// r_InetInfoSetAdminInformationResponse carries the [out] parameters and return value of R_InetInfoSetAdminInformation.
type r_InetInfoSetAdminInformationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_InetInfoSetAdminInformation calls R_InetInfoSetAdminInformation (opnum 3) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoSetAdminInformation(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD, pConfig msirp.INET_INFO_CONFIG_INFO) (err error) {
	req := &r_InetInfoSetAdminInformationRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
		PConfig:      pConfig,
	}
	var resp r_InetInfoSetAdminInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoSetAdminInformation: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoSetAdminInformation failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
