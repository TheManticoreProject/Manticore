package functions

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumSubnetClientsRequest carries the [in] parameters of R_DhcpEnumSubnetClients.
type r_DhcpEnumSubnetClientsRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    ndr.DWORD
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpEnumSubnetClientsRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpEnumSubnetClients }

// r_DhcpEnumSubnetClientsResponse carries the [out] parameters and return value of R_DhcpEnumSubnetClients.
type r_DhcpEnumSubnetClientsResponse struct {
	ResumeHandle ndr.DWORD
	ClientInfo   *msdhcpm.DHCP_CLIENT_INFO_ARRAY `ndr:"unique"`
	ClientsRead  ndr.DWORD
	ClientsTotal ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumSubnetClients calls R_DhcpEnumSubnetClients (opnum 20) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumSubnetClients(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, ClientInfo *msdhcpm.DHCP_CLIENT_INFO_ARRAY, ClientsRead ndr.DWORD, ClientsTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumSubnetClientsRequest{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpEnumSubnetClientsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumSubnetClients: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	ClientInfo = resp.ClientInfo
	ClientsRead = resp.ClientsRead
	ClientsTotal = resp.ClientsTotal
	if uint32(resp.Status) != dhcpsrv.StatusSuccess && !dhcpsrv.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumSubnetClients failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
