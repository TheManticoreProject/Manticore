package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpDeleteClassRequest carries the [in] parameters of R_DhcpDeleteClass.
type r_DhcpDeleteClassRequest struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	ClassName          *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpDeleteClassRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteClass }

// r_DhcpDeleteClassResponse carries the [out] parameters and return value of R_DhcpDeleteClass.
type r_DhcpDeleteClassResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteClass calls R_DhcpDeleteClass (opnum 26) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteClass(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, className *ndr.WSTR) (err error) {
	req := &r_DhcpDeleteClassRequest{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		ClassName:          className,
	}
	var resp r_DhcpDeleteClassResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteClass: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteClass failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
