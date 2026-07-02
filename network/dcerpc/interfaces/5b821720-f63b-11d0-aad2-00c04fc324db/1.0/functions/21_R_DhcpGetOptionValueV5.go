package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetOptionValueV5Request carries the [in] parameters of R_DhcpGetOptionValueV5.
type r_DhcpGetOptionValueV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpGetOptionValueV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetOptionValueV5 }

// r_DhcpGetOptionValueV5Response carries the [out] parameters and return value of R_DhcpGetOptionValueV5.
type r_DhcpGetOptionValueV5Response struct {
	OptionValue *msdhcpm.DHCP_OPTION_VALUE `ndr:"unique"`
	Status      ndr.DWORD                  `ndr:"retval"`
}

// R_DhcpGetOptionValueV5 calls R_DhcpGetOptionValueV5 (opnum 21) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetOptionValueV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (OptionValue *msdhcpm.DHCP_OPTION_VALUE, err error) {
	req := &r_DhcpGetOptionValueV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		ClassName:       className,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpGetOptionValueV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetOptionValueV5: %w", err)
		return
	}
	OptionValue = resp.OptionValue
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetOptionValueV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
