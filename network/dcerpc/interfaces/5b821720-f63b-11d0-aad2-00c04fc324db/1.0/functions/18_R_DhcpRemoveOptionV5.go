package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpRemoveOptionV5Request carries the [in] parameters of R_DhcpRemoveOptionV5.
type r_DhcpRemoveOptionV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpRemoveOptionV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpRemoveOptionV5 }

// r_DhcpRemoveOptionV5Response carries the [out] parameters and return value of R_DhcpRemoveOptionV5.
type r_DhcpRemoveOptionV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRemoveOptionV5 calls R_DhcpRemoveOptionV5 (opnum 18) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRemoveOptionV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR) (err error) {
	req := &r_DhcpRemoveOptionV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		ClassName:       className,
		VendorName:      vendorName,
	}
	var resp r_DhcpRemoveOptionV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRemoveOptionV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpRemoveOptionV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
