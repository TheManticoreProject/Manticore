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

// r_DhcpEnumSubnetElementsV5Request carries the [in] parameters of R_DhcpEnumSubnetElementsV5.
type r_DhcpEnumSubnetElementsV5Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    ndr.DWORD
	EnumElementType  msdhcpm.DHCP_SUBNET_ELEMENT_TYPE
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumSubnetElementsV5Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpEnumSubnetElementsV5
}

// r_DhcpEnumSubnetElementsV5Response carries the [out] parameters and return value of R_DhcpEnumSubnetElementsV5.
type r_DhcpEnumSubnetElementsV5Response struct {
	ResumeHandle    ndr.DWORD
	EnumElementInfo *msdhcpm.DHCP_SUBNET_ELEMENT_INFO_ARRAY_V5 `ndr:"unique"`
	ElementsRead    ndr.DWORD
	ElementsTotal   ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumSubnetElementsV5 calls R_DhcpEnumSubnetElementsV5 (opnum 38) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumSubnetElementsV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, enumElementType msdhcpm.DHCP_SUBNET_ELEMENT_TYPE, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, EnumElementInfo *msdhcpm.DHCP_SUBNET_ELEMENT_INFO_ARRAY_V5, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumSubnetElementsV5Request{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		EnumElementType:  enumElementType,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumSubnetElementsV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumSubnetElementsV5: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumElementInfo = resp.EnumElementInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumSubnetElementsV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
