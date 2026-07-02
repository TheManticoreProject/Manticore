package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumSubnetClientsV6Request carries the [in] parameters of R_DhcpEnumSubnetClientsV6.
type r_DhcpEnumSubnetClientsV6Request struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    msdhcpm.DHCP_IPV6_ADDRESS
	ResumeHandle     msdhcpm.DHCP_RESUME_IPV6_HANDLE
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumSubnetClientsV6Request) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpEnumSubnetClientsV6
}

// r_DhcpEnumSubnetClientsV6Response carries the [out] parameters and return value of R_DhcpEnumSubnetClientsV6.
type r_DhcpEnumSubnetClientsV6Response struct {
	ResumeHandle msdhcpm.DHCP_RESUME_IPV6_HANDLE
	ClientInfo   *msdhcpm.DHCP_CLIENT_INFO_ARRAY_V6 `ndr:"unique"`
	ClientsRead  ndr.DWORD
	ClientsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumSubnetClientsV6 calls R_DhcpEnumSubnetClientsV6 (opnum 64) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumSubnetClientsV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress msdhcpm.DHCP_IPV6_ADDRESS, resumeHandle msdhcpm.DHCP_RESUME_IPV6_HANDLE, preferredMaximum ndr.DWORD) (ResumeHandle msdhcpm.DHCP_RESUME_IPV6_HANDLE, ClientInfo *msdhcpm.DHCP_CLIENT_INFO_ARRAY_V6, ClientsRead ndr.DWORD, ClientsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumSubnetClientsV6Request{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumSubnetClientsV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumSubnetClientsV6: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	ClientInfo = resp.ClientInfo
	ClientsRead = resp.ClientsRead
	ClientsTotal = resp.ClientsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumSubnetClientsV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
