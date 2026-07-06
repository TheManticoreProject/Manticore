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

// r_DhcpGetServerSpecificStringsRequest carries the [in] parameters of R_DhcpGetServerSpecificStrings.
type r_DhcpGetServerSpecificStringsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetServerSpecificStringsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpGetServerSpecificStrings
}

// r_DhcpGetServerSpecificStringsResponse carries the [out] parameters and return value of R_DhcpGetServerSpecificStrings.
type r_DhcpGetServerSpecificStringsResponse struct {
	ServerSpecificStrings *msdhcpm.DHCP_SERVER_SPECIFIC_STRINGS `ndr:"unique"`
	Status                ndr.DWORD                             `ndr:"retval"`
}

// R_DhcpGetServerSpecificStrings calls R_DhcpGetServerSpecificStrings (opnum 46) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetServerSpecificStrings(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (ServerSpecificStrings *msdhcpm.DHCP_SERVER_SPECIFIC_STRINGS, err error) {
	req := &r_DhcpGetServerSpecificStringsRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetServerSpecificStringsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetServerSpecificStrings: %w", err)
		return
	}
	ServerSpecificStrings = resp.ServerSpecificStrings
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetServerSpecificStrings failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
