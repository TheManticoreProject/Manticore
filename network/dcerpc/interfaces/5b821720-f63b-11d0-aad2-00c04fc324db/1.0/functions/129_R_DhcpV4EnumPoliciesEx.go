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

// r_DhcpV4EnumPoliciesExRequest carries the [in] parameters of R_DhcpV4EnumPoliciesEx.
type r_DhcpV4EnumPoliciesExRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
	ServerPolicy     ndr.BOOL
	SubnetAddress    ndr.DWORD
}

func (*r_DhcpV4EnumPoliciesExRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpV4EnumPoliciesEx }

// r_DhcpV4EnumPoliciesExResponse carries the [out] parameters and return value of R_DhcpV4EnumPoliciesEx.
type r_DhcpV4EnumPoliciesExResponse struct {
	ResumeHandle  ndr.DWORD
	EnumInfo      msdhcpm.DHCP_POLICY_EX_ARRAY
	ElementsRead  ndr.DWORD
	ElementsTotal ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4EnumPoliciesEx calls R_DhcpV4EnumPoliciesEx (opnum 129) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4EnumPoliciesEx(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD, serverPolicy ndr.BOOL, subnetAddress ndr.DWORD) (ResumeHandle ndr.DWORD, EnumInfo msdhcpm.DHCP_POLICY_EX_ARRAY, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpV4EnumPoliciesExRequest{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
		ServerPolicy:     serverPolicy,
		SubnetAddress:    subnetAddress,
	}
	var resp r_DhcpV4EnumPoliciesExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4EnumPoliciesEx: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumInfo = resp.EnumInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpV4EnumPoliciesEx failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
