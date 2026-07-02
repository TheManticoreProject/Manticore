package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpDeleteClassV6Request carries the [in] parameters of R_DhcpDeleteClassV6.
type r_DhcpDeleteClassV6Request struct {
	ServerIpAddress    *ndr.WSTR `ndr:"unique"`
	ReservedMustBeZero ndr.DWORD
	ClassName          *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpDeleteClassV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpDeleteClassV6 }

// r_DhcpDeleteClassV6Response carries the [out] parameters and return value of R_DhcpDeleteClassV6.
type r_DhcpDeleteClassV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpDeleteClassV6 calls R_DhcpDeleteClassV6 (opnum 76) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpDeleteClassV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, reservedMustBeZero ndr.DWORD, className *ndr.WSTR) (err error) {
	req := &r_DhcpDeleteClassV6Request{
		ServerIpAddress:    serverIpAddress,
		ReservedMustBeZero: reservedMustBeZero,
		ClassName:          className,
	}
	var resp r_DhcpDeleteClassV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpDeleteClassV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpDeleteClassV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
