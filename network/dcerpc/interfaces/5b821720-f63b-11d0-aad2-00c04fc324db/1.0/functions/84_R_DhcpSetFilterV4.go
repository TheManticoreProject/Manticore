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

// r_DhcpSetFilterV4Request carries the [in] parameters of R_DhcpSetFilterV4.
type r_DhcpSetFilterV4Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	GlobalFilterInfo msdhcpm.DHCP_FILTER_GLOBAL_INFO
}

func (*r_DhcpSetFilterV4Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetFilterV4 }

// r_DhcpSetFilterV4Response carries the [out] parameters and return value of R_DhcpSetFilterV4.
type r_DhcpSetFilterV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetFilterV4 calls R_DhcpSetFilterV4 (opnum 84) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetFilterV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, globalFilterInfo msdhcpm.DHCP_FILTER_GLOBAL_INFO) (err error) {
	req := &r_DhcpSetFilterV4Request{
		ServerIpAddress:  serverIpAddress,
		GlobalFilterInfo: globalFilterInfo,
	}
	var resp r_DhcpSetFilterV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetFilterV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetFilterV4 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
