package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetClientInfoVQRequest carries the [in] parameters of R_DhcpGetClientInfoVQ.
type r_DhcpGetClientInfoVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SearchInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpGetClientInfoVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetClientInfoVQ }

// r_DhcpGetClientInfoVQResponse carries the [out] parameters and return value of R_DhcpGetClientInfoVQ.
type r_DhcpGetClientInfoVQResponse struct {
	ClientInfo *msdhcpm.DHCP_CLIENT_INFO_VQ `ndr:"unique"`
	Status     ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpGetClientInfoVQ calls R_DhcpGetClientInfoVQ (opnum 46) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetClientInfoVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, searchInfo msdhcpm.DHCP_SEARCH_INFO) (ClientInfo *msdhcpm.DHCP_CLIENT_INFO_VQ, err error) {
	req := &r_DhcpGetClientInfoVQRequest{
		ServerIpAddress: serverIpAddress,
		SearchInfo:      searchInfo,
	}
	var resp r_DhcpGetClientInfoVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetClientInfoVQ: %w", err)
		return
	}
	ClientInfo = resp.ClientInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetClientInfoVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
