package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumMScopesRequest carries the [in] parameters of R_DhcpEnumMScopes.
type r_DhcpEnumMScopesRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumMScopesRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpEnumMScopes }

// r_DhcpEnumMScopesResponse carries the [out] parameters and return value of R_DhcpEnumMScopes.
type r_DhcpEnumMScopesResponse struct {
	ResumeHandle  ndr.DWORD
	MScopeTable   *msdhcpm.DHCP_MSCOPE_TABLE `ndr:"unique"`
	ElementsRead  ndr.DWORD
	ElementsTotal ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumMScopes calls R_DhcpEnumMScopes (opnum 3) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumMScopes(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, MScopeTable *msdhcpm.DHCP_MSCOPE_TABLE, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumMScopesRequest{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumMScopesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumMScopes: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	MScopeTable = resp.MScopeTable
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumMScopes failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
