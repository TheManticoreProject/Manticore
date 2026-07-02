package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpCreateMClientInfoRequest carries the [in] parameters of R_DhcpCreateMClientInfo.
type r_DhcpCreateMClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	MScopeName      ndr.WSTR
	ClientInfo      msdhcpm.DHCP_MCLIENT_INFO
}

func (*r_DhcpCreateMClientInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpCreateMClientInfo }

// r_DhcpCreateMClientInfoResponse carries the [out] parameters and return value of R_DhcpCreateMClientInfo.
type r_DhcpCreateMClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateMClientInfo calls R_DhcpCreateMClientInfo (opnum 9) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateMClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, mScopeName ndr.WSTR, clientInfo msdhcpm.DHCP_MCLIENT_INFO) (err error) {
	req := &r_DhcpCreateMClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		MScopeName:      mScopeName,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpCreateMClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateMClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateMClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
