package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerGetConfigVQRequest carries the [in] parameters of R_DhcpServerGetConfigVQ.
type r_DhcpServerGetConfigVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpServerGetConfigVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpServerGetConfigVQ }

// r_DhcpServerGetConfigVQResponse carries the [out] parameters and return value of R_DhcpServerGetConfigVQ.
type r_DhcpServerGetConfigVQResponse struct {
	ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO_VQ `ndr:"unique"`
	Status     ndr.DWORD                           `ndr:"retval"`
}

// R_DhcpServerGetConfigVQ calls R_DhcpServerGetConfigVQ (opnum 42) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerGetConfigVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO_VQ, err error) {
	req := &r_DhcpServerGetConfigVQRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpServerGetConfigVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerGetConfigVQ: %w", err)
		return
	}
	ConfigInfo = resp.ConfigInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerGetConfigVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
