package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpDeleteClientInfoRequest carries the [in] parameters of R_DhcpDeleteClientInfo.
type r_DhcpDeleteClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpDeleteClientInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpDeleteClientInfo }

// r_DhcpDeleteClientInfoResponse carries the [out] parameters and return value of R_DhcpDeleteClientInfo.
type r_DhcpDeleteClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteClientInfo calls R_DhcpDeleteClientInfo (opnum 19) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_SEARCH_INFO) (err error) {
	req := &r_DhcpDeleteClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpDeleteClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteClientInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
