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

// r_DhcpGetClientInfoV6Request carries the [in] parameters of R_DhcpGetClientInfoV6.
type r_DhcpGetClientInfoV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SearchInfo      msdhcpm.DHCP_SEARCH_INFO_V6
}

func (*r_DhcpGetClientInfoV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetClientInfoV6 }

// r_DhcpGetClientInfoV6Response carries the [out] parameters and return value of R_DhcpGetClientInfoV6.
type r_DhcpGetClientInfoV6Response struct {
	ClientInfo *msdhcpm.DHCP_CLIENT_INFO_V6 `ndr:"unique"`
	Status     ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpGetClientInfoV6 calls R_DhcpGetClientInfoV6 (opnum 72) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetClientInfoV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, searchInfo msdhcpm.DHCP_SEARCH_INFO_V6) (ClientInfo *msdhcpm.DHCP_CLIENT_INFO_V6, err error) {
	req := &r_DhcpGetClientInfoV6Request{
		ServerIpAddress: serverIpAddress,
		SearchInfo:      searchInfo,
	}
	var resp r_DhcpGetClientInfoV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetClientInfoV6: %w", err)
		return
	}
	ClientInfo = resp.ClientInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetClientInfoV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
