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

// r_DhcpDeleteMClientInfoRequest carries the [in] parameters of R_DhcpDeleteMClientInfo.
type r_DhcpDeleteMClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpDeleteMClientInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteMClientInfo }

// r_DhcpDeleteMClientInfoResponse carries the [out] parameters and return value of R_DhcpDeleteMClientInfo.
type r_DhcpDeleteMClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteMClientInfo calls R_DhcpDeleteMClientInfo (opnum 12) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteMClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_SEARCH_INFO) (err error) {
	req := &r_DhcpDeleteMClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpDeleteMClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteMClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteMClientInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
