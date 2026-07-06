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

// r_DhcpRemoveOptionValueRequest carries the [in] parameters of R_DhcpRemoveOptionValue.
type r_DhcpRemoveOptionValueRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	OptionID        ndr.DWORD
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
}

func (*r_DhcpRemoveOptionValueRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpRemoveOptionValue }

// r_DhcpRemoveOptionValueResponse carries the [out] parameters and return value of R_DhcpRemoveOptionValue.
type r_DhcpRemoveOptionValueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveOptionValue calls R_DhcpRemoveOptionValue (opnum 15) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveOptionValue(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, optionID ndr.DWORD, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO) (err error) {
	req := &r_DhcpRemoveOptionValueRequest{
		ServerIpAddress: serverIpAddress,
		OptionID:        optionID,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpRemoveOptionValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveOptionValue: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveOptionValue failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
