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

// r_DhcpGetMibInfoVQRequest carries the [in] parameters of R_DhcpGetMibInfoVQ.
type r_DhcpGetMibInfoVQRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetMibInfoVQRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetMibInfoVQ }

// r_DhcpGetMibInfoVQResponse carries the [out] parameters and return value of R_DhcpGetMibInfoVQ.
type r_DhcpGetMibInfoVQResponse struct {
	MibInfo *msdhcpm.DHCP_MIB_INFO_VQ `ndr:"unique"`
	Status  ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetMibInfoVQ calls R_DhcpGetMibInfoVQ (opnum 43) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetMibInfoVQ(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (MibInfo *msdhcpm.DHCP_MIB_INFO_VQ, err error) {
	req := &r_DhcpGetMibInfoVQRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetMibInfoVQResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetMibInfoVQ: %w", err)
		return
	}
	MibInfo = resp.MibInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetMibInfoVQ failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
