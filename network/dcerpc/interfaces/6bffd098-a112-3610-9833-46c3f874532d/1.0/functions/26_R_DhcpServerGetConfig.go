package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerGetConfigRequest carries the [in] parameters of R_DhcpServerGetConfig.
type r_DhcpServerGetConfigRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpServerGetConfigRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpServerGetConfig }

// r_DhcpServerGetConfigResponse carries the [out] parameters and return value of R_DhcpServerGetConfig.
type r_DhcpServerGetConfigResponse struct {
	ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO `ndr:"unique"`
	Status     ndr.DWORD                        `ndr:"retval"`
}

// R_DhcpServerGetConfig calls R_DhcpServerGetConfig (opnum 26) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerGetConfig(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO, err error) {
	req := &r_DhcpServerGetConfigRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpServerGetConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerGetConfig: %w", err)
		return
	}
	ConfigInfo = resp.ConfigInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerGetConfig failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
