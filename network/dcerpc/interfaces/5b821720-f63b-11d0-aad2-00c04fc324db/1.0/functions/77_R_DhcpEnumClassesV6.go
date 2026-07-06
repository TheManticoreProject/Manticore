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

// r_DhcpEnumClassesV6Request carries the [in] parameters of R_DhcpEnumClassesV6.
type r_DhcpEnumClassesV6Request struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	ResumeHandle       ndr.DWORD
	PreferredMaximum   ndr.DWORD
}

func (*r_DhcpEnumClassesV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpEnumClassesV6 }

// r_DhcpEnumClassesV6Response carries the [out] parameters and return value of R_DhcpEnumClassesV6.
type r_DhcpEnumClassesV6Response struct {
	ResumeHandle   ndr.DWORD
	ClassInfoArray *msdhcpm.DHCP_CLASS_INFO_ARRAY_V6 `ndr:"unique"`
	NRead          ndr.DWORD
	NTotal         ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// R_DhcpEnumClassesV6 calls R_DhcpEnumClassesV6 (opnum 77) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpEnumClassesV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, ClassInfoArray *msdhcpm.DHCP_CLASS_INFO_ARRAY_V6, NRead ndr.DWORD, NTotal ndr.DWORD, err error) {
	req := &r_DhcpEnumClassesV6Request{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		ResumeHandle:       resumeHandle,
		PreferredMaximum:   preferredMaximum,
	}
	var resp r_DhcpEnumClassesV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpEnumClassesV6: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	ClassInfoArray = resp.ClassInfoArray
	NRead = resp.NRead
	NTotal = resp.NTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpEnumClassesV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
