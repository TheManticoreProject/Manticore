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

// r_DhcpSetMScopeInfoRequest carries the [in] parameters of R_DhcpSetMScopeInfo.
type r_DhcpSetMScopeInfoRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	MScopeName      ndr.WSTR
	MScopeInfo      msdhcpm.DHCP_MSCOPE_INFO
	NewScope        ndr.BOOL
}

func (*r_DhcpSetMScopeInfoRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetMScopeInfo }

// r_DhcpSetMScopeInfoResponse carries the [out] parameters and return value of R_DhcpSetMScopeInfo.
type r_DhcpSetMScopeInfoResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetMScopeInfo calls R_DhcpSetMScopeInfo (opnum 1) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetMScopeInfo(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, mScopeName ndr.WSTR, mScopeInfo msdhcpm.DHCP_MSCOPE_INFO, newScope ndr.BOOL) (err error) {
	req := &r_DhcpSetMScopeInfoRequest{
		ServerIpAddress: serverIpAddress,
		MScopeName:      mScopeName,
		MScopeInfo:      mScopeInfo,
		NewScope:        newScope,
	}
	var resp r_DhcpSetMScopeInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetMScopeInfo: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetMScopeInfo failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
