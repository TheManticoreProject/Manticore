package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetFilterV4Request carries the [in] parameters of R_DhcpGetFilterV4.
type r_DhcpGetFilterV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetFilterV4Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetFilterV4 }

// r_DhcpGetFilterV4Response carries the [out] parameters and return value of R_DhcpGetFilterV4.
type r_DhcpGetFilterV4Response struct {
	GlobalFilterInfo msdhcpm.DHCP_FILTER_GLOBAL_INFO
	Status           ndr.DWORD `ndr:"retval"`
}

// R_DhcpGetFilterV4 calls R_DhcpGetFilterV4 (opnum 85) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetFilterV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (GlobalFilterInfo msdhcpm.DHCP_FILTER_GLOBAL_INFO, err error) {
	req := &r_DhcpGetFilterV4Request{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetFilterV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetFilterV4: %w", err)
		return
	}
	GlobalFilterInfo = resp.GlobalFilterInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetFilterV4 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
