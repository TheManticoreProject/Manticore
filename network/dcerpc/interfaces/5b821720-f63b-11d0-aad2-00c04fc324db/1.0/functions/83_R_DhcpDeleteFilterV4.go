package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpDeleteFilterV4Request carries the [in] parameters of R_DhcpDeleteFilterV4.
type r_DhcpDeleteFilterV4Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	DeleteFilterInfo msdhcpm.DHCP_ADDR_PATTERN
}

func (*r_DhcpDeleteFilterV4Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteFilterV4 }

// r_DhcpDeleteFilterV4Response carries the [out] parameters and return value of R_DhcpDeleteFilterV4.
type r_DhcpDeleteFilterV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteFilterV4 calls R_DhcpDeleteFilterV4 (opnum 83) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteFilterV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, deleteFilterInfo msdhcpm.DHCP_ADDR_PATTERN) (err error) {
	req := &r_DhcpDeleteFilterV4Request{
		ServerIpAddress:  serverIpAddress,
		DeleteFilterInfo: deleteFilterInfo,
	}
	var resp r_DhcpDeleteFilterV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteFilterV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteFilterV4 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
