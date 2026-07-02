package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpModifyClassRequest carries the [in] parameters of R_DhcpModifyClass.
type r_DhcpModifyClassRequest struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	ClassInfo          msdhcpm.DHCP_CLASS_INFO
}

func (*r_DhcpModifyClassRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpModifyClass }

// r_DhcpModifyClassResponse carries the [out] parameters and return value of R_DhcpModifyClass.
type r_DhcpModifyClassResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpModifyClass calls R_DhcpModifyClass (opnum 25) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpModifyClass(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, classInfo msdhcpm.DHCP_CLASS_INFO) (err error) {
	req := &r_DhcpModifyClassRequest{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		ClassInfo:          classInfo,
	}
	var resp r_DhcpModifyClassResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpModifyClass: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpModifyClass failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
