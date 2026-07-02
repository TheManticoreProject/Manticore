package functions

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msirp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-irp"
)

// r_InetInfoGetSitesRequest carries the [in] parameters of R_InetInfoGetSites.
type r_InetInfoGetSitesRequest struct {
	PszServer    *ndr.WSTR `ndr:"unique"`
	DwServerMask ndr.DWORD
}

func (*r_InetInfoGetSitesRequest) Opnum() uint16 { return inetinfo.OpnumR_InetInfoGetSites }

// r_InetInfoGetSitesResponse carries the [out] parameters and return value of R_InetInfoGetSites.
type r_InetInfoGetSitesResponse struct {
	PpSites *msirp.INET_INFO_SITE_LIST `ndr:"unique"`
	Status  ndr.DWORD                  `ndr:"retval"`
}

// R_InetInfoGetSites calls R_InetInfoGetSites (opnum 2) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_InetInfoGetSites(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServerMask ndr.DWORD) (PpSites *msirp.INET_INFO_SITE_LIST, err error) {
	req := &r_InetInfoGetSitesRequest{
		PszServer:    pszServer,
		DwServerMask: dwServerMask,
	}
	var resp r_InetInfoGetSitesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_InetInfoGetSites: %w", err)
		return
	}
	PpSites = resp.PpSites
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_InetInfoGetSites failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
