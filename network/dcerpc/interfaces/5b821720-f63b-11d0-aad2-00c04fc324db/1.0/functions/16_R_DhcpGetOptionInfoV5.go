package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetOptionInfoV5Request carries the [in] parameters of R_DhcpGetOptionInfoV5.
type r_DhcpGetOptionInfoV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetOptionInfoV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetOptionInfoV5 }

// r_DhcpGetOptionInfoV5Response carries the [out] parameters and return value of R_DhcpGetOptionInfoV5.
type r_DhcpGetOptionInfoV5Response struct {
	OptionInfo *msdhcpm.DHCP_OPTION `ndr:"unique"`
	Status     ndr.DWORD            `ndr:"retval"`
}

// R_DhcpGetOptionInfoV5 calls R_DhcpGetOptionInfoV5 (opnum 16) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetOptionInfoV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR) (OptionInfo *msdhcpm.DHCP_OPTION, err error) {
	req := &r_DhcpGetOptionInfoV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		ClassName:       className,
		VendorName:      vendorName,
	}
	var resp r_DhcpGetOptionInfoV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetOptionInfoV5: %w", err)
		return
	}
	OptionInfo = resp.OptionInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetOptionInfoV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
