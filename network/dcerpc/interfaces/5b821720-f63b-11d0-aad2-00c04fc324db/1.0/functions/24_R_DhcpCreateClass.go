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

// r_DhcpCreateClassRequest carries the [in] parameters of R_DhcpCreateClass.
type r_DhcpCreateClassRequest struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	ClassInfo          msdhcpm.DHCP_CLASS_INFO
}

func (*r_DhcpCreateClassRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpCreateClass }

// r_DhcpCreateClassResponse carries the [out] parameters and return value of R_DhcpCreateClass.
type r_DhcpCreateClassResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpCreateClass calls R_DhcpCreateClass (opnum 24) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpCreateClass(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, classInfo msdhcpm.DHCP_CLASS_INFO) (err error) {
	req := &r_DhcpCreateClassRequest{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		ClassInfo:          classInfo,
	}
	var resp r_DhcpCreateClassResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpCreateClass: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpCreateClass failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
