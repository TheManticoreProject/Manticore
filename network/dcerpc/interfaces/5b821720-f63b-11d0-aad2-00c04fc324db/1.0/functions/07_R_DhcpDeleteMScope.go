package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpDeleteMScopeRequest carries the [in] parameters of R_DhcpDeleteMScope.
type r_DhcpDeleteMScopeRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	MScopeName      ndr.WSTR
	ForceFlag       msdhcpm.DHCP_FORCE_FLAG
}

func (*r_DhcpDeleteMScopeRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteMScope }

// r_DhcpDeleteMScopeResponse carries the [out] parameters and return value of R_DhcpDeleteMScope.
type r_DhcpDeleteMScopeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteMScope calls R_DhcpDeleteMScope (opnum 7) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteMScope(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, mScopeName ndr.WSTR, forceFlag msdhcpm.DHCP_FORCE_FLAG) (err error) {
	req := &r_DhcpDeleteMScopeRequest{
		ServerIpAddress: serverIpAddress,
		MScopeName:      mScopeName,
		ForceFlag:       forceFlag,
	}
	var resp r_DhcpDeleteMScopeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteMScope: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteMScope failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
