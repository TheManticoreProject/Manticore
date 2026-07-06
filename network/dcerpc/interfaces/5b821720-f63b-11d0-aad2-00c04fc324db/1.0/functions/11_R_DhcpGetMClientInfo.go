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

// r_DhcpGetMClientInfoRequest carries the [in] parameters of R_DhcpGetMClientInfo.
type r_DhcpGetMClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SearchInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpGetMClientInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetMClientInfo }

// r_DhcpGetMClientInfoResponse carries the [out] parameters and return value of R_DhcpGetMClientInfo.
type r_DhcpGetMClientInfoResponse struct {
	ClientInfo *msdhcpm.DHCP_MCLIENT_INFO `ndr:"unique"`
	Status     ndr.DWORD                  `ndr:"retval"`
}

// R_DhcpGetMClientInfo calls R_DhcpGetMClientInfo (opnum 11) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetMClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, searchInfo msdhcpm.DHCP_SEARCH_INFO) (ClientInfo *msdhcpm.DHCP_MCLIENT_INFO, err error) {
	req := &r_DhcpGetMClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		SearchInfo:      searchInfo,
	}
	var resp r_DhcpGetMClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetMClientInfo: %w", err)
		return
	}
	ClientInfo = resp.ClientInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetMClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
