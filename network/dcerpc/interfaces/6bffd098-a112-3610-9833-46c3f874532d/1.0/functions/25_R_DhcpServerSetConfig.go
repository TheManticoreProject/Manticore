package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerSetConfigRequest carries the [in] parameters of R_DhcpServerSetConfig.
type r_DhcpServerSetConfigRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FieldsToSet     ndr.DWORD
	ConfigInfo      msdhcpm.DHCP_SERVER_CONFIG_INFO
}

func (*r_DhcpServerSetConfigRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpServerSetConfig }

// r_DhcpServerSetConfigResponse carries the [out] parameters and return value of R_DhcpServerSetConfig.
type r_DhcpServerSetConfigResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpServerSetConfig calls R_DhcpServerSetConfig (opnum 25) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerSetConfig(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fieldsToSet ndr.DWORD, configInfo msdhcpm.DHCP_SERVER_CONFIG_INFO) (err error) {
	req := &r_DhcpServerSetConfigRequest{
		ServerIpAddress: serverIpAddress,
		FieldsToSet:     fieldsToSet,
		ConfigInfo:      configInfo,
	}
	var resp r_DhcpServerSetConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerSetConfig: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerSetConfig failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
