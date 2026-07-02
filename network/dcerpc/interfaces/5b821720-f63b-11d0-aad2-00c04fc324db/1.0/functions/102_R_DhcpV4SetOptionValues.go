package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4SetOptionValuesRequest carries the [in] parameters of R_DhcpV4SetOptionValues.
type r_DhcpV4SetOptionValuesRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
	OptionValues    msdhcpm.DHCP_OPTION_VALUE_ARRAY
}

func (*r_DhcpV4SetOptionValuesRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4SetOptionValues }

// r_DhcpV4SetOptionValuesResponse carries the [out] parameters and return value of R_DhcpV4SetOptionValues.
type r_DhcpV4SetOptionValuesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4SetOptionValues calls R_DhcpV4SetOptionValues (opnum 102) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4SetOptionValues(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, policyName *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO, optionValues msdhcpm.DHCP_OPTION_VALUE_ARRAY) (err error) {
	req := &r_DhcpV4SetOptionValuesRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		PolicyName:      policyName,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
		OptionValues:    optionValues,
	}
	var resp r_DhcpV4SetOptionValuesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4SetOptionValues: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4SetOptionValues failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
