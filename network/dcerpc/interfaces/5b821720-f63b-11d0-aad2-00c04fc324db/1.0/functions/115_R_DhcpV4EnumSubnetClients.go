package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4EnumSubnetClientsRequest carries the [in] parameters of R_DhcpV4EnumSubnetClients.
type r_DhcpV4EnumSubnetClientsRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    ndr.DWORD
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpV4EnumSubnetClientsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4EnumSubnetClients
}

// r_DhcpV4EnumSubnetClientsResponse carries the [out] parameters and return value of R_DhcpV4EnumSubnetClients.
type r_DhcpV4EnumSubnetClientsResponse struct {
	ResumeHandle ndr.DWORD
	ClientInfo   *msdhcpm.DHCP_CLIENT_INFO_PB_ARRAY `ndr:"unique"`
	ClientsRead  ndr.DWORD
	ClientsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4EnumSubnetClients calls R_DhcpV4EnumSubnetClients (opnum 115) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4EnumSubnetClients(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, ClientInfo *msdhcpm.DHCP_CLIENT_INFO_PB_ARRAY, ClientsRead ndr.DWORD, ClientsTotal ndr.DWORD, err error) {
	req := &r_DhcpV4EnumSubnetClientsRequest{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpV4EnumSubnetClientsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4EnumSubnetClients: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	ClientInfo = resp.ClientInfo
	ClientsRead = resp.ClientsRead
	ClientsTotal = resp.ClientsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpV4EnumSubnetClients failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
