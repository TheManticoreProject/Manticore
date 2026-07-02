package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetMCastMibInfoRequest carries the [in] parameters of R_DhcpGetMCastMibInfo.
type r_DhcpGetMCastMibInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetMCastMibInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetMCastMibInfo }

// r_DhcpGetMCastMibInfoResponse carries the [out] parameters and return value of R_DhcpGetMCastMibInfo.
type r_DhcpGetMCastMibInfoResponse struct {
	MibInfo *msdhcpm.DHCP_MCAST_MIB_INFO `ndr:"unique"`
	Status  ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpGetMCastMibInfo calls R_DhcpGetMCastMibInfo (opnum 31) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetMCastMibInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (MibInfo *msdhcpm.DHCP_MCAST_MIB_INFO, err error) {
	req := &r_DhcpGetMCastMibInfoRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetMCastMibInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetMCastMibInfo: %w", err)
		return
	}
	MibInfo = resp.MibInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetMCastMibInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
