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

// r_DhcpEnumSubnetElementsV6Request carries the [in] parameters of R_DhcpEnumSubnetElementsV6.
type r_DhcpEnumSubnetElementsV6Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    msdhcpm.DHCP_IPV6_ADDRESS
	EnumElementType  msdhcpm.DHCP_SUBNET_ELEMENT_TYPE_V6
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumSubnetElementsV6Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpEnumSubnetElementsV6
}

// r_DhcpEnumSubnetElementsV6Response carries the [out] parameters and return value of R_DhcpEnumSubnetElementsV6.
type r_DhcpEnumSubnetElementsV6Response struct {
	ResumeHandle    ndr.DWORD
	EnumElementInfo *msdhcpm.DHCP_SUBNET_ELEMENT_INFO_ARRAY_V6 `ndr:"unique"`
	ElementsRead    ndr.DWORD
	ElementsTotal   ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumSubnetElementsV6 calls R_DhcpEnumSubnetElementsV6 (opnum 60) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumSubnetElementsV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, enumElementType msdhcpm.DHCP_SUBNET_ELEMENT_TYPE_V6, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, EnumElementInfo *msdhcpm.DHCP_SUBNET_ELEMENT_INFO_ARRAY_V6, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumSubnetElementsV6Request{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		EnumElementType:  enumElementType,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumSubnetElementsV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumSubnetElementsV6: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumElementInfo = resp.EnumElementInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumSubnetElementsV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
