package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetOptionInfoV6Request carries the [in] parameters of R_DhcpSetOptionInfoV6.
type r_DhcpSetOptionInfoV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	OptionInfo      msdhcpm.DHCP_OPTION
}

func (*r_DhcpSetOptionInfoV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetOptionInfoV6 }

// r_DhcpSetOptionInfoV6Response carries the [out] parameters and return value of R_DhcpSetOptionInfoV6.
type r_DhcpSetOptionInfoV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionInfoV6 calls R_DhcpSetOptionInfoV6 (opnum 48) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, optionInfo msdhcpm.DHCP_OPTION) (err error) {
	req := &r_DhcpSetOptionInfoV6Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		ClassName:       className,
		VendorName:      vendorName,
		OptionInfo:      optionInfo,
	}
	var resp r_DhcpSetOptionInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionInfoV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
