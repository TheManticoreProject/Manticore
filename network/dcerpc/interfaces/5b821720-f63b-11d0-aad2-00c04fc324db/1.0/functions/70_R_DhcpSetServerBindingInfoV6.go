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

// r_DhcpSetServerBindingInfoV6Request carries the [in] parameters of R_DhcpSetServerBindingInfoV6.
type r_DhcpSetServerBindingInfoV6Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	Flags            ndr.DWORD
	BindElementsInfo msdhcpm.DHCPV6_BIND_ELEMENT_ARRAY
}

func (*r_DhcpSetServerBindingInfoV6Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpSetServerBindingInfoV6
}

// r_DhcpSetServerBindingInfoV6Response carries the [out] parameters and return value of R_DhcpSetServerBindingInfoV6.
type r_DhcpSetServerBindingInfoV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetServerBindingInfoV6 calls R_DhcpSetServerBindingInfoV6 (opnum 70) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetServerBindingInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, bindElementsInfo msdhcpm.DHCPV6_BIND_ELEMENT_ARRAY) (err error) {
	req := &r_DhcpSetServerBindingInfoV6Request{
		ServerIpAddress:  serverIpAddress,
		Flags:            flags,
		BindElementsInfo: bindElementsInfo,
	}
	var resp r_DhcpSetServerBindingInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetServerBindingInfoV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetServerBindingInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
