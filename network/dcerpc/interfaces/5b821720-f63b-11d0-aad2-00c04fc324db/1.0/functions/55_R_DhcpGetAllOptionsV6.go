package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetAllOptionsV6Request carries the [in] parameters of R_DhcpGetAllOptionsV6.
type r_DhcpGetAllOptionsV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
}

func (*r_DhcpGetAllOptionsV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetAllOptionsV6 }

// r_DhcpGetAllOptionsV6Response carries the [out] parameters and return value of R_DhcpGetAllOptionsV6.
type r_DhcpGetAllOptionsV6Response struct {
	OptionStruct *msdhcpm.DHCP_ALL_OPTIONS `ndr:"unique"`
	Status       ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetAllOptionsV6 calls R_DhcpGetAllOptionsV6 (opnum 55) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetAllOptionsV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD) (OptionStruct *msdhcpm.DHCP_ALL_OPTIONS, err error) {
	req := &r_DhcpGetAllOptionsV6Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
	}
	var resp r_DhcpGetAllOptionsV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetAllOptionsV6: %w", err)
		return
	}
	OptionStruct = resp.OptionStruct
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetAllOptionsV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
