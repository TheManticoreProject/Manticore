package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumFilterV4Request carries the [in] parameters of R_DhcpEnumFilterV4.
type r_DhcpEnumFilterV4Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     msdhcpm.DHCP_ADDR_PATTERN
	PreferredMaximum ndr.DWORD
	ListType         msdhcpm.DHCP_FILTER_LIST_TYPE
}

func (*r_DhcpEnumFilterV4Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpEnumFilterV4 }

// r_DhcpEnumFilterV4Response carries the [out] parameters and return value of R_DhcpEnumFilterV4.
type r_DhcpEnumFilterV4Response struct {
	ResumeHandle   msdhcpm.DHCP_ADDR_PATTERN
	EnumFilterInfo *msdhcpm.DHCP_FILTER_ENUM_INFO `ndr:"unique"`
	ElementsRead   ndr.DWORD
	ElementsTotal  ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumFilterV4 calls R_DhcpEnumFilterV4 (opnum 86) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumFilterV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle msdhcpm.DHCP_ADDR_PATTERN, preferredMaximum ndr.DWORD, listType msdhcpm.DHCP_FILTER_LIST_TYPE) (ResumeHandle msdhcpm.DHCP_ADDR_PATTERN, EnumFilterInfo *msdhcpm.DHCP_FILTER_ENUM_INFO, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumFilterV4Request{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
		ListType:         listType,
	}
	var resp r_DhcpEnumFilterV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumFilterV4: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumFilterInfo = resp.EnumFilterInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumFilterV4 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
