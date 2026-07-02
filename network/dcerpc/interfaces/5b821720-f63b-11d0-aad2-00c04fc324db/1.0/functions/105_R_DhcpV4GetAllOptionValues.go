package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4GetAllOptionValuesRequest carries the [in] parameters of R_DhcpV4GetAllOptionValues.
type r_DhcpV4GetAllOptionValuesRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpV4GetAllOptionValuesRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4GetAllOptionValues
}

// r_DhcpV4GetAllOptionValuesResponse carries the [out] parameters and return value of R_DhcpV4GetAllOptionValues.
type r_DhcpV4GetAllOptionValuesResponse struct {
	Values *msdhcpm.DHCP_ALL_OPTION_VALUES_PB `ndr:"unique"`
	Status ndr.DWORD                          `ndr:"retval"`
}

// R_DhcpV4GetAllOptionValues calls R_DhcpV4GetAllOptionValues (opnum 105) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4GetAllOptionValues(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (Values *msdhcpm.DHCP_ALL_OPTION_VALUES_PB, err error) {
	req := &r_DhcpV4GetAllOptionValuesRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpV4GetAllOptionValuesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4GetAllOptionValues: %w", err)
		return
	}
	Values = resp.Values
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4GetAllOptionValues failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
