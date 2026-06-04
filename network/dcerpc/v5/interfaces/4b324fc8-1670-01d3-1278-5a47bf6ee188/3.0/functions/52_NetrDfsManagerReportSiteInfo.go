package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
)

// netrDfsManagerReportSiteInfoRequest is the [in] parameter set of
// NetrDfsManagerReportSiteInfo: the [unique] server name and the [in,out,unique] pointer
// to the DFS site list (an LPDFS_SITELIST_INFO* in the IDL, modelled as a [unique]
// pointer present in both request and response).
type netrDfsManagerReportSiteInfoRequest struct {
	ServerName *ndr.WSTR                     `ndr:"unique"`
	PpSiteInfo *structures.DFS_SITELIST_INFO `ndr:"unique"`
}

func (*netrDfsManagerReportSiteInfoRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrDfsManagerReportSiteInfo
}

// netrDfsManagerReportSiteInfoResponse is the reply: the [in,out,unique] site list and
// the NET_API_STATUS return value.
type netrDfsManagerReportSiteInfoResponse struct {
	PpSiteInfo *structures.DFS_SITELIST_INFO `ndr:"unique"`
	Status     ndr.DWORD                     `ndr:"retval"`
}

// NetrDfsManagerReportSiteInfo calls NetrDfsManagerReportSiteInfo (opnum 52), retrieving
// the list of sites known to the DFS server ([MS-SRVS] 3.1.4.52).
func NetrDfsManagerReportSiteInfo(rpc *client.Client, serverName string, ppSiteInfo *structures.DFS_SITELIST_INFO) (*structures.DFS_SITELIST_INFO, error) {
	req := &netrDfsManagerReportSiteInfoRequest{
		ServerName: optWStr(serverName),
		PpSiteInfo: ppSiteInfo,
	}
	var resp netrDfsManagerReportSiteInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrDfsManagerReportSiteInfo: %w", err)
	}
	status := uint32(resp.Status)
	if status != srvsvc.NERR_Success && status != srvsvc.ERROR_MORE_DATA {
		return resp.PpSiteInfo, fmt.Errorf("NetrDfsManagerReportSiteInfo failed: %s", srvsvc.StatusString(status))
	}
	return resp.PpSiteInfo, nil
}
