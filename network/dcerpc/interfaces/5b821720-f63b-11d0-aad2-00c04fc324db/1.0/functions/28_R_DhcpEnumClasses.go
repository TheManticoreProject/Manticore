package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpEnumClassesRequest carries the [in] parameters of R_DhcpEnumClasses.
type r_DhcpEnumClassesRequest struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	ResumeHandle       ndr.DWORD
	PreferredMaximum   ndr.DWORD
}

func (*r_DhcpEnumClassesRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpEnumClasses }

// r_DhcpEnumClassesResponse carries the [out] parameters and return value of R_DhcpEnumClasses.
type r_DhcpEnumClassesResponse struct {
	ResumeHandle   ndr.DWORD
	ClassInfoArray *msdhcpm.DHCP_CLASS_INFO_ARRAY `ndr:"unique"`
	NRead          ndr.DWORD
	NTotal         ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumClasses calls R_DhcpEnumClasses (opnum 28) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumClasses(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, ClassInfoArray *msdhcpm.DHCP_CLASS_INFO_ARRAY, NRead ndr.DWORD, NTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumClassesRequest{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		ResumeHandle:       resumeHandle,
		PreferredMaximum:   preferredMaximum,
	}
	var resp r_DhcpEnumClassesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumClasses: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	ClassInfoArray = resp.ClassInfoArray
	NRead = resp.NRead
	NTotal = resp.NTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumClasses failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
