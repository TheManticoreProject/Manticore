package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpCreateOptionV6Request carries the [in] parameters of R_DhcpCreateOptionV6.
type r_DhcpCreateOptionV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionId        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	OptionInfo      msdhcpm.DHCP_OPTION
}

func (*r_DhcpCreateOptionV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpCreateOptionV6 }

// r_DhcpCreateOptionV6Response carries the [out] parameters and return value of R_DhcpCreateOptionV6.
type r_DhcpCreateOptionV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateOptionV6 calls R_DhcpCreateOptionV6 (opnum 47) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateOptionV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionId ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, optionInfo msdhcpm.DHCP_OPTION) (err error) {
	req := &r_DhcpCreateOptionV6Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionId:        optionId,
		ClassName:       className,
		VendorName:      vendorName,
		OptionInfo:      optionInfo,
	}
	var resp r_DhcpCreateOptionV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateOptionV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateOptionV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
