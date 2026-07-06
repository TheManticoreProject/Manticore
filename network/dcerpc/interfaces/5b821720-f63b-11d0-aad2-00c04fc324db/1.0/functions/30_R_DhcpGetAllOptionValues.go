package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpGetAllOptionValuesRequest carries the [in] parameters of R_DhcpGetAllOptionValues.
type r_DhcpGetAllOptionValuesRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpGetAllOptionValuesRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetAllOptionValues }

// r_DhcpGetAllOptionValuesResponse carries the [out] parameters and return value of R_DhcpGetAllOptionValues.
type r_DhcpGetAllOptionValuesResponse struct {
	Values *msdhcpm.DHCP_ALL_OPTION_VALUES `ndr:"unique"`
	Status ndr.DWORD                       `ndr:"retval"`
}

// R_DhcpGetAllOptionValues calls R_DhcpGetAllOptionValues (opnum 30) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetAllOptionValues(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (Values *msdhcpm.DHCP_ALL_OPTION_VALUES, err error) {
	req := &r_DhcpGetAllOptionValuesRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpGetAllOptionValuesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetAllOptionValues: %w", err)
		return
	}
	Values = resp.Values
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetAllOptionValues failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
