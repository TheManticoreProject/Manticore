package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4EnumPoliciesRequest carries the [in] parameters of R_DhcpV4EnumPolicies.
type r_DhcpV4EnumPoliciesRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
	ServerPolicy     ndr.BOOL
	SubnetAddress    ndr.DWORD
}

func (*r_DhcpV4EnumPoliciesRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4EnumPolicies }

// r_DhcpV4EnumPoliciesResponse carries the [out] parameters and return value of R_DhcpV4EnumPolicies.
type r_DhcpV4EnumPoliciesResponse struct {
	ResumeHandle  ndr.DWORD
	EnumInfo      msdhcpm.DHCP_POLICY_ARRAY
	ElementsRead  ndr.DWORD
	ElementsTotal ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4EnumPolicies calls R_DhcpV4EnumPolicies (opnum 112) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4EnumPolicies(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD) (ResumeHandle ndr.DWORD, EnumInfo msdhcpm.DHCP_POLICY_ARRAY, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpV4EnumPoliciesRequest{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
		ServerPolicy:     serverPolicy,
		SubnetAddress:    subnetAddress,
	}
	var resp r_DhcpV4EnumPoliciesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4EnumPolicies: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumInfo = resp.EnumInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpV4EnumPolicies failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
