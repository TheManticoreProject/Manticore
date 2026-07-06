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

// r_DhcpSetClientInfoVQRequest carries the [in] parameters of R_DhcpSetClientInfoVQ.
type r_DhcpSetClientInfoVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ClientInfo      msdhcpm.DHCP_CLIENT_INFO_VQ
}

func (*r_DhcpSetClientInfoVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpSetClientInfoVQ }

// r_DhcpSetClientInfoVQResponse carries the [out] parameters and return value of R_DhcpSetClientInfoVQ.
type r_DhcpSetClientInfoVQResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetClientInfoVQ calls R_DhcpSetClientInfoVQ (opnum 45) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetClientInfoVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, clientInfo msdhcpm.DHCP_CLIENT_INFO_VQ) (err error) {
	req := &r_DhcpSetClientInfoVQRequest{
		ServerIpAddress: serverIpAddress,
		ClientInfo:      clientInfo,
	}
	var resp r_DhcpSetClientInfoVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetClientInfoVQ: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetClientInfoVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
