package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetMibInfoRequest carries the [in] parameters of R_DhcpGetMibInfo.
type r_DhcpGetMibInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetMibInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetMibInfo }

// r_DhcpGetMibInfoResponse carries the [out] parameters and return value of R_DhcpGetMibInfo.
type r_DhcpGetMibInfoResponse struct {
	MibInfo *msdhcpm.DHCP_MIB_INFO `ndr:"unique"`
	Status  ndr.DWORD              `ndr:"retval"`
}

// R_DhcpGetMibInfo calls R_DhcpGetMibInfo (opnum 22) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetMibInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (MibInfo *msdhcpm.DHCP_MIB_INFO, err error) {
	req := &r_DhcpGetMibInfoRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetMibInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetMibInfo: %w", err)
		return
	}
	MibInfo = resp.MibInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetMibInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
