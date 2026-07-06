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

// r_DhcpGetClientInfoV4Request carries the [in] parameters of R_DhcpGetClientInfoV4.
type r_DhcpGetClientInfoV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	SearchInfo      msdhcpm.DHCP_SEARCH_INFO
}

func (*r_DhcpGetClientInfoV4Request) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetClientInfoV4 }

// r_DhcpGetClientInfoV4Response carries the [out] parameters and return value of R_DhcpGetClientInfoV4.
type r_DhcpGetClientInfoV4Response struct {
	ClientInfo *msdhcpm.DHCP_CLIENT_INFO_V4 `ndr:"unique"`
	Status     ndr.DWORD                    `ndr:"retval"`
}

// R_DhcpGetClientInfoV4 calls R_DhcpGetClientInfoV4 (opnum 34) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetClientInfoV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, searchInfo msdhcpm.DHCP_SEARCH_INFO) (ClientInfo *msdhcpm.DHCP_CLIENT_INFO_V4, err error) {
	req := &r_DhcpGetClientInfoV4Request{
		ServerIpAddress: serverIpAddress,
		SearchInfo:      searchInfo,
	}
	var resp r_DhcpGetClientInfoV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetClientInfoV4: %w", err)
		return
	}
	ClientInfo = resp.ClientInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetClientInfoV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
