package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpRemoveOptionValueV5Request carries the [in] parameters of R_DhcpRemoveOptionValueV5.
type r_DhcpRemoveOptionValueV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpRemoveOptionValueV5Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpRemoveOptionValueV5
}

// r_DhcpRemoveOptionValueV5Response carries the [out] parameters and return value of R_DhcpRemoveOptionValueV5.
type r_DhcpRemoveOptionValueV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveOptionValueV5 calls R_DhcpRemoveOptionValueV5 (opnum 23) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveOptionValueV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (err error) {
	req := &r_DhcpRemoveOptionValueV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		ClassName:       className,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpRemoveOptionValueV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveOptionValueV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveOptionValueV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
