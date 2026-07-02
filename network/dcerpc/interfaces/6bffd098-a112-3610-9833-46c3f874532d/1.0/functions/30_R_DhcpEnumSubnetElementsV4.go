package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumSubnetElementsV4Request carries the [in] parameters of R_DhcpEnumSubnetElementsV4.
type r_DhcpEnumSubnetElementsV4Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    ndr.DWORD
	EnumElementType  msdhcpm.DHCP_SUBNET_ELEMENT_TYPE
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumSubnetElementsV4Request) Opnum() uint16 {
	return dhcpsrv.OpnumR_DhcpEnumSubnetElementsV4
}

// r_DhcpEnumSubnetElementsV4Response carries the [out] parameters and return value of R_DhcpEnumSubnetElementsV4.
type r_DhcpEnumSubnetElementsV4Response struct {
	ResumeHandle    ndr.DWORD
	EnumElementInfo *msdhcpm.DHCP_SUBNET_ELEMENT_INFO_ARRAY_V4 `ndr:"unique"`
	ElementsRead    ndr.DWORD
	ElementsTotal   ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumSubnetElementsV4 calls R_DhcpEnumSubnetElementsV4 (opnum 30) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumSubnetElementsV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, enumElementType msdhcpm.DHCP_SUBNET_ELEMENT_TYPE, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, EnumElementInfo *msdhcpm.DHCP_SUBNET_ELEMENT_INFO_ARRAY_V4, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumSubnetElementsV4Request{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		EnumElementType:  enumElementType,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumSubnetElementsV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumSubnetElementsV4: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumElementInfo = resp.EnumElementInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv.StatusSuccess && !dhcpsrv.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumSubnetElementsV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
