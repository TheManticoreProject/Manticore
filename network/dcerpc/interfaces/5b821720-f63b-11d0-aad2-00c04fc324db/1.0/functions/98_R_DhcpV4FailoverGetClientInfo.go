package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4FailoverGetClientInfoRequest carries the [in] parameters of R_DhcpV4FailoverGetClientInfo.
type r_DhcpV4FailoverGetClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SearchInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpV4FailoverGetClientInfoRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverGetClientInfo
}

// r_DhcpV4FailoverGetClientInfoResponse carries the [out] parameters and return value of R_DhcpV4FailoverGetClientInfo.
type r_DhcpV4FailoverGetClientInfoResponse struct {
	ClientInfo *msdhcpm.DHCPV4_FAILOVER_CLIENT_INFO `ndr:"unique"`
	Status     ndr.DWORD                            `ndr:"retval"`
}

// R_DhcpV4FailoverGetClientInfo calls R_DhcpV4FailoverGetClientInfo (opnum 98) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverGetClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, searchInfo msdhcpm.DHCP_SEARCH_INFO) (ClientInfo *msdhcpm.DHCPV4_FAILOVER_CLIENT_INFO, err error) {
	req := &r_DhcpV4FailoverGetClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		SearchInfo:      searchInfo,
	}
	var resp r_DhcpV4FailoverGetClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverGetClientInfo: %w", err)
		return
	}
	ClientInfo = resp.ClientInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4FailoverGetClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
