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

// r_DhcpGetAllOptionValuesV6Request carries the [in] parameters of R_DhcpGetAllOptionValuesV6.
type r_DhcpGetAllOptionValuesV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO6
}

func (*r_DhcpGetAllOptionValuesV6Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpGetAllOptionValuesV6
}

// r_DhcpGetAllOptionValuesV6Response carries the [out] parameters and return value of R_DhcpGetAllOptionValuesV6.
type r_DhcpGetAllOptionValuesV6Response struct {
	Values *msdhcpm.DHCP_ALL_OPTION_VALUES `ndr:"unique"`
	Status ndr.DWORD                       `ndr:"retval"`
}

// R_DhcpGetAllOptionValuesV6 calls R_DhcpGetAllOptionValuesV6 (opnum 56) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetAllOptionValuesV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO6) (Values *msdhcpm.DHCP_ALL_OPTION_VALUES, err error) {
	req := &r_DhcpGetAllOptionValuesV6Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		ScopeInfo:       scopeInfo,
	}
	var resp r_DhcpGetAllOptionValuesV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetAllOptionValuesV6: %w", err)
		return
	}
	Values = resp.Values
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetAllOptionValuesV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
