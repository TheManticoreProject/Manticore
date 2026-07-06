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

// r_DhcpAddFilterV4Request carries the [in] parameters of R_DhcpAddFilterV4.
type r_DhcpAddFilterV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	AddFilterInfo   msdhcpm.DHCP_FILTER_ADD_INFO
	ForceFlag       ndr.BOOL
}

func (*r_DhcpAddFilterV4Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpAddFilterV4 }

// r_DhcpAddFilterV4Response carries the [out] parameters and return value of R_DhcpAddFilterV4.
type r_DhcpAddFilterV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpAddFilterV4 calls R_DhcpAddFilterV4 (opnum 82) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpAddFilterV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, addFilterInfo msdhcpm.DHCP_FILTER_ADD_INFO, forceFlag ndr.BOOL) (err error) {
	req := &r_DhcpAddFilterV4Request{
		ServerIpAddress: serverIpAddress,
		AddFilterInfo:   addFilterInfo,
		ForceFlag:       forceFlag,
	}
	var resp r_DhcpAddFilterV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpAddFilterV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpAddFilterV4 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
