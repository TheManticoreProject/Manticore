package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4RemoveOptionValueRequest carries the [in] parameters of R_DhcpV4RemoveOptionValue.
type r_DhcpV4RemoveOptionValueRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	PolicyName      *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpV4RemoveOptionValueRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4RemoveOptionValue
}

// r_DhcpV4RemoveOptionValueResponse carries the [out] parameters and return value of R_DhcpV4RemoveOptionValue.
type r_DhcpV4RemoveOptionValueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4RemoveOptionValue calls R_DhcpV4RemoveOptionValue (opnum 104) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4RemoveOptionValue(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, policyName *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (err error) {
	req := &r_DhcpV4RemoveOptionValueRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		PolicyName:      policyName,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpV4RemoveOptionValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4RemoveOptionValue: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4RemoveOptionValue failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
