package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4EnumSubnetReservationsRequest carries the [in] parameters of R_DhcpV4EnumSubnetReservations.
type r_DhcpV4EnumSubnetReservationsRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	SubnetAddress    ndr.DWORD
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpV4EnumSubnetReservationsRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4EnumSubnetReservations
}

// r_DhcpV4EnumSubnetReservationsResponse carries the [out] parameters and return value of R_DhcpV4EnumSubnetReservations.
type r_DhcpV4EnumSubnetReservationsResponse struct {
	ResumeHandle    ndr.DWORD
	EnumElementInfo msdhcpm.DHCP_RESERVATION_INFO_ARRAY
	ElementsRead    ndr.DWORD
	ElementsTotal   ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4EnumSubnetReservations calls R_DhcpV4EnumSubnetReservations (opnum 119) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4EnumSubnetReservations(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, subnetAddress ndr.DWORD, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, EnumElementInfo msdhcpm.DHCP_RESERVATION_INFO_ARRAY, ElementsRead ndr.DWORD, ElementsTotal ndr.DWORD, err error) {
	req := &r_DhcpV4EnumSubnetReservationsRequest{
		ServerIpAddress:  serverIpAddress,
		SubnetAddress:    subnetAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpV4EnumSubnetReservationsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4EnumSubnetReservations: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	EnumElementInfo = resp.EnumElementInfo
	ElementsRead = resp.ElementsRead
	ElementsTotal = resp.ElementsTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpV4EnumSubnetReservations failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
