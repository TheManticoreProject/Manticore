package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4GetOptionValueRequest carries the [in] parameters of R_DhcpV4GetOptionValue.
type r_DhcpV4GetOptionValueRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpV4GetOptionValueRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4GetOptionValue }

// r_DhcpV4GetOptionValueResponse carries the [out] parameters and return value of R_DhcpV4GetOptionValue.
type r_DhcpV4GetOptionValueResponse struct {
	OptionValue *msdhcpm.DHCP_OPTION_VALUE `ndr:"unique"`
	Status      ndr.DWORD                  `ndr:"retval"`
}

// R_DhcpV4GetOptionValue calls R_DhcpV4GetOptionValue (opnum 103) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4GetOptionValue(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, policyName *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (OptionValue *msdhcpm.DHCP_OPTION_VALUE, err error) {
	req := &r_DhcpV4GetOptionValueRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		PolicyName:      policyName,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpV4GetOptionValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4GetOptionValue: %w", err)
		return
	}
	OptionValue = resp.OptionValue
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4GetOptionValue failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
