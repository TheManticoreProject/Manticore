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

// r_DhcpGetServerBindingInfoV6Request carries the [in] parameters of R_DhcpGetServerBindingInfoV6.
type r_DhcpGetServerBindingInfoV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
}

func (*r_DhcpGetServerBindingInfoV6Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpGetServerBindingInfoV6
}

// r_DhcpGetServerBindingInfoV6Response carries the [out] parameters and return value of R_DhcpGetServerBindingInfoV6.
type r_DhcpGetServerBindingInfoV6Response struct {
	BindElementsInfo *msdhcpm.DHCPV6_BIND_ELEMENT_ARRAY `ndr:"unique"`
	Status           ndr.DWORD                          `ndr:"retval"`
}

// R_DhcpGetServerBindingInfoV6 calls R_DhcpGetServerBindingInfoV6 (opnum 69) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetServerBindingInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD) (BindElementsInfo *msdhcpm.DHCPV6_BIND_ELEMENT_ARRAY, err error) {
	req := &r_DhcpGetServerBindingInfoV6Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
	}
	var resp r_DhcpGetServerBindingInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetServerBindingInfoV6: %w", err)
		return
	}
	BindElementsInfo = resp.BindElementsInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetServerBindingInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
