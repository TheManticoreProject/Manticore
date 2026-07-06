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

// r_DhcpEnumSubnetsRequest carries the [in] parameters of R_DhcpEnumSubnets.
type r_DhcpEnumSubnetsRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumSubnetsRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpEnumSubnets }

// r_DhcpEnumSubnetsResponse carries the [out] parameters and return value of R_DhcpEnumSubnets.
type r_DhcpEnumSubnetsResponse struct {
	ResumeHandle  ndr.DWORD
	EnumInfo      *msdhcpm.DHCP_IP_ARRAY `ndr:"unique"`
	ElementsRead  ndr.DWORD
	ElementsTotal ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumSubnets calls R_DhcpEnumSubnets (opnum 3) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumSubnets(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, EnumInfo *msdhcpm.DHCP_IP_ARRAY, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumSubnetsRequest{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumSubnetsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumSubnets: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumInfo = resp.EnumInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv.StatusSuccess && !dhcpsrv.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumSubnets failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
