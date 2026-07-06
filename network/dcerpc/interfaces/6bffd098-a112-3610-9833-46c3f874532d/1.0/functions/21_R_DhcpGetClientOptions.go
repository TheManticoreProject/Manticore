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

// r_DhcpGetClientOptionsRequest carries the [in] parameters of R_DhcpGetClientOptions.
type r_DhcpGetClientOptionsRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ClientIpAddress  ndr.DWORD
	ClientSubnetMask ndr.DWORD
}

func (*r_DhcpGetClientOptionsRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetClientOptions }

// r_DhcpGetClientOptionsResponse carries the [out] parameters and return value of R_DhcpGetClientOptions.
type r_DhcpGetClientOptionsResponse struct {
	ClientOptions *msdhcpm.DHCP_OPTION_LIST `ndr:"unique"`
	Status        ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetClientOptions calls R_DhcpGetClientOptions (opnum 21) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetClientOptions(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientIpAddress ndr.DWORD, clientSubnetMask ndr.DWORD) (ClientOptions *msdhcpm.DHCP_OPTION_LIST, err error) {
	req := &r_DhcpGetClientOptionsRequest{
		ServerIpAddress:  serverIpAddress,
		ClientIpAddress:  clientIpAddress,
		ClientSubnetMask: clientSubnetMask,
	}
	var resp r_DhcpGetClientOptionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetClientOptions: %w", err)
		return
	}
	ClientOptions = resp.ClientOptions
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetClientOptions failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
