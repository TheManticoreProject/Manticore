package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerGetConfigV6Request carries the [in] parameters of R_DhcpServerGetConfigV6.
type r_DhcpServerGetConfigV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO6
}

func (*r_DhcpServerGetConfigV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpServerGetConfigV6 }

// r_DhcpServerGetConfigV6Response carries the [out] parameters and return value of R_DhcpServerGetConfigV6.
type r_DhcpServerGetConfigV6Response struct {
	ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO_V6 `ndr:"unique"`
	Status     ndr.DWORD                           `ndr:"retval"`
}

// R_DhcpServerGetConfigV6 calls R_DhcpServerGetConfigV6 (opnum 66) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerGetConfigV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO6) (ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO_V6, err error) {
	req := &r_DhcpServerGetConfigV6Request{
		ServerIpAddress: serverIpAddress,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpServerGetConfigV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerGetConfigV6: %w", err)
		return
	}
	ConfigInfo = resp.ConfigInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerGetConfigV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
