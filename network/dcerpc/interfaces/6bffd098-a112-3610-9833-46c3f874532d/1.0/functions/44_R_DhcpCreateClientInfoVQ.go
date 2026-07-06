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

// r_DhcpCreateClientInfoVQRequest carries the [in] parameters of R_DhcpCreateClientInfoVQ.
type r_DhcpCreateClientInfoVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO_VQ
}

func (*r_DhcpCreateClientInfoVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpCreateClientInfoVQ }

// r_DhcpCreateClientInfoVQResponse carries the [out] parameters and return value of R_DhcpCreateClientInfoVQ.
type r_DhcpCreateClientInfoVQResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateClientInfoVQ calls R_DhcpCreateClientInfoVQ (opnum 44) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateClientInfoVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO_VQ) (err error) {
	req := &r_DhcpCreateClientInfoVQRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpCreateClientInfoVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateClientInfoVQ: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateClientInfoVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
