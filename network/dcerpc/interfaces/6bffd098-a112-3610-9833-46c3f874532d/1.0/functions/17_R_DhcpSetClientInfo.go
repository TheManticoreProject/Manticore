package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetClientInfoRequest carries the [in] parameters of R_DhcpSetClientInfo.
type r_DhcpSetClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO
}

func (*r_DhcpSetClientInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetClientInfo }

// r_DhcpSetClientInfoResponse carries the [out] parameters and return value of R_DhcpSetClientInfo.
type r_DhcpSetClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetClientInfo calls R_DhcpSetClientInfo (opnum 17) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO) (err error) {
	req := &r_DhcpSetClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpSetClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetClientInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
