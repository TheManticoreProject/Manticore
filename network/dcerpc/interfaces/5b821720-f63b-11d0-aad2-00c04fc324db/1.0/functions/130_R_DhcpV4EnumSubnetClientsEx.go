package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4EnumSubnetClientsExRequest carries the [in] parameters of R_DhcpV4EnumSubnetClientsEx.
type r_DhcpV4EnumSubnetClientsExRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    ndr.DWORD
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpV4EnumSubnetClientsExRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4EnumSubnetClientsEx
}

// r_DhcpV4EnumSubnetClientsExResponse carries the [out] parameters and return value of R_DhcpV4EnumSubnetClientsEx.
type r_DhcpV4EnumSubnetClientsExResponse struct {
	ResumeHandle ndr.DWORD
	ClientInfo   *msdhcpm.DHCP_CLIENT_INFO_EX_ARRAY `ndr:"unique"`
	ClientsRead  ndr.DWORD
	ClientsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4EnumSubnetClientsEx calls R_DhcpV4EnumSubnetClientsEx (opnum 130) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4EnumSubnetClientsEx(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, ClientInfo *msdhcpm.DHCP_CLIENT_INFO_EX_ARRAY, ClientsRead ndr.DWORD, ClientsTotal ndr.DWORD, err error) {
	req := &r_DhcpV4EnumSubnetClientsExRequest{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpV4EnumSubnetClientsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4EnumSubnetClientsEx: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	ClientInfo = resp.ClientInfo
	ClientsRead = resp.ClientsRead
	ClientsTotal = resp.ClientsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpV4EnumSubnetClientsEx failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
