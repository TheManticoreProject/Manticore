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

// r_DhcpGetMScopeInfoRequest carries the [in] parameters of R_DhcpGetMScopeInfo.
type r_DhcpGetMScopeInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	MScopeName      ndr.WSTR
}

func (*r_DhcpGetMScopeInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetMScopeInfo }

// r_DhcpGetMScopeInfoResponse carries the [out] parameters and return value of R_DhcpGetMScopeInfo.
type r_DhcpGetMScopeInfoResponse struct {
	MScopeInfo *msdhcpm.DHCP_MSCOPE_INFO `ndr:"unique"`
	Status     ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetMScopeInfo calls R_DhcpGetMScopeInfo (opnum 2) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetMScopeInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, mScopeName ndr.WSTR) (MScopeInfo *msdhcpm.DHCP_MSCOPE_INFO, err error) {
	req := &r_DhcpGetMScopeInfoRequest{
		ServerIpAddress: serverIpAddress,
		MScopeName:      mScopeName,
	}
	var resp r_DhcpGetMScopeInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetMScopeInfo: %w", err)
		return
	}
	MScopeInfo = resp.MScopeInfo
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetMScopeInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
