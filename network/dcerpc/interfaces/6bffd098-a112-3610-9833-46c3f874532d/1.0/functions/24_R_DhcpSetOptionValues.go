package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpSetOptionValuesRequest carries the [in] parameters of R_DhcpSetOptionValues.
type r_DhcpSetOptionValuesRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
	OptionValues    msdhcpm.DHCP_OPTION_VALUE_ARRAY
}

func (*r_DhcpSetOptionValuesRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetOptionValues }

// r_DhcpSetOptionValuesResponse carries the [out] parameters and return value of R_DhcpSetOptionValues.
type r_DhcpSetOptionValuesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionValues calls R_DhcpSetOptionValues (opnum 24) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionValues(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO, optionValues msdhcpm.DHCP_OPTION_VALUE_ARRAY) (err error) {
	req := &r_DhcpSetOptionValuesRequest{
		ServerIpAddress: serverIpAddress,
		ScopeInfo:       scopeInfo,
		OptionValues:    optionValues,
	}
	var resp r_DhcpSetOptionValuesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionValues: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionValues failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
