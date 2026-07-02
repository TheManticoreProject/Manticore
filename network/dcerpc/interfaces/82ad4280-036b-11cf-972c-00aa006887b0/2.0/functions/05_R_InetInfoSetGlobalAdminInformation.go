package functions

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// r_InetInfoSetGlobalAdminInformationRequest carries the [in] parameters of R_InetInfoSetGlobalAdminInformation.
type r_InetInfoSetGlobalAdminInformationRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
	PConfig      msirp.INET_INFO_GLOBAL_CONFIG_INFO
}

func (*r_InetInfoSetGlobalAdminInformationRequest) Opnum() uint16 {
	return inetinfo.OpnumR_InetInfoSetGlobalAdminInformation
}

// r_InetInfoSetGlobalAdminInformationResponse carries the [out] parameters and return value of R_InetInfoSetGlobalAdminInformation.
type r_InetInfoSetGlobalAdminInformationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_InetInfoSetGlobalAdminInformation calls R_InetInfoSetGlobalAdminInformation (opnum 5) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoSetGlobalAdminInformation(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD, pConfig msirp.INET_INFO_GLOBAL_CONFIG_INFO) (err error) {
	req := &r_InetInfoSetGlobalAdminInformationRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
		PConfig:      pConfig,
	}
	var resp r_InetInfoSetGlobalAdminInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoSetGlobalAdminInformation: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoSetGlobalAdminInformation failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
