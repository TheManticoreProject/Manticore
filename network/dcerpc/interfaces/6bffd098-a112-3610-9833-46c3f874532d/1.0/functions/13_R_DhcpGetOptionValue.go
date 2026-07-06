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

// r_DhcpGetOptionValueRequest carries the [in] parameters of R_DhcpGetOptionValue.
type r_DhcpGetOptionValueRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	OptionID        ndr.DWORD
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpGetOptionValueRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetOptionValue }

// r_DhcpGetOptionValueResponse carries the [out] parameters and return value of R_DhcpGetOptionValue.
type r_DhcpGetOptionValueResponse struct {
	OptionValue *msdhcpm.DHCP_OPTION_VALUE `ndr:"unique"`
	Status      ndr.DWORD                  `ndr:"retval"`
}

// R_DhcpGetOptionValue calls R_DhcpGetOptionValue (opnum 13) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetOptionValue(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, optionID ndr.DWORD, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (OptionValue *msdhcpm.DHCP_OPTION_VALUE, err error) {
	req := &r_DhcpGetOptionValueRequest{
		ServerIpAddress: serverIpAddress,
		OptionID:        optionID,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpGetOptionValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetOptionValue: %w", err)
		return
	}
	OptionValue = resp.OptionValue
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetOptionValue failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
