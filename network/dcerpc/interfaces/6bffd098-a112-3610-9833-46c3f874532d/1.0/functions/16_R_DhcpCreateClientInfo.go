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

// r_DhcpCreateClientInfoRequest carries the [in] parameters of R_DhcpCreateClientInfo.
type r_DhcpCreateClientInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO
}

func (*r_DhcpCreateClientInfoRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpCreateClientInfo }

// r_DhcpCreateClientInfoResponse carries the [out] parameters and return value of R_DhcpCreateClientInfo.
type r_DhcpCreateClientInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateClientInfo calls R_DhcpCreateClientInfo (opnum 16) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateClientInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO) (err error) {
	req := &r_DhcpCreateClientInfoRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpCreateClientInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateClientInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateClientInfo failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
