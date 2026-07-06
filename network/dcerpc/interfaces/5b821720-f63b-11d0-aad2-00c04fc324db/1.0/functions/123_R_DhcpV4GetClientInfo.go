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

// r_DhcpV4GetClientInfoRequest carries the [in] parameters of R_DhcpV4GetClientInfo.
type r_DhcpV4GetClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SearchInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpV4GetClientInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4GetClientInfo }

// r_DhcpV4GetClientInfoResponse carries the [out] parameters and return value of R_DhcpV4GetClientInfo.
type r_DhcpV4GetClientInfoResponse struct {
	ClientInfo *msdhcpm.DHCP_CLIENT_INFO_PB `ndr:"unique"`
	Status     ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpV4GetClientInfo calls R_DhcpV4GetClientInfo (opnum 123) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4GetClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, searchInfo msdhcpm.DHCP_SEARCH_INFO) (ClientInfo *msdhcpm.DHCP_CLIENT_INFO_PB, err error) {
	req := &r_DhcpV4GetClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		SearchInfo:      searchInfo,
	}
	var resp r_DhcpV4GetClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4GetClientInfo: %w", err)
		return
	}
	ClientInfo = resp.ClientInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpV4GetClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
