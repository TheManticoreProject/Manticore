package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetClientInfoV4Request carries the [in] parameters of R_DhcpSetClientInfoV4.
type r_DhcpSetClientInfoV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO_V4
}

func (*r_DhcpSetClientInfoV4Request) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetClientInfoV4 }

// r_DhcpSetClientInfoV4Response carries the [out] parameters and return value of R_DhcpSetClientInfoV4.
type r_DhcpSetClientInfoV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetClientInfoV4 calls R_DhcpSetClientInfoV4 (opnum 33) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetClientInfoV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO_V4) (err error) {
	req := &r_DhcpSetClientInfoV4Request{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpSetClientInfoV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetClientInfoV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetClientInfoV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
