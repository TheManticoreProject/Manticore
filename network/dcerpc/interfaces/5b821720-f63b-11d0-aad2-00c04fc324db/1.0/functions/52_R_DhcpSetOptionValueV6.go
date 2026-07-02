package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetOptionValueV6Request carries the [in] parameters of R_DhcpSetOptionValueV6.
type r_DhcpSetOptionValueV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionId        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO6
	OptionValue     msdhcpm.DHCP_OPTION_DATA
}

func (*r_DhcpSetOptionValueV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetOptionValueV6 }

// r_DhcpSetOptionValueV6Response carries the [out] parameters and return value of R_DhcpSetOptionValueV6.
type r_DhcpSetOptionValueV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionValueV6 calls R_DhcpSetOptionValueV6 (opnum 52) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionValueV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionId ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO6, optionValue msdhcpm.DHCP_OPTION_DATA) (err error) {
	req := &r_DhcpSetOptionValueV6Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionId:        optionId,
		ClassName:       className,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
		OptionValue:     optionValue,
	}
	var resp r_DhcpSetOptionValueV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionValueV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionValueV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
