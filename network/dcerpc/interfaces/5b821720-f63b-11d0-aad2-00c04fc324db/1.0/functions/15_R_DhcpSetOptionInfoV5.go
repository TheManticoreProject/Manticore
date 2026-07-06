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

// r_DhcpSetOptionInfoV5Request carries the [in] parameters of R_DhcpSetOptionInfoV5.
type r_DhcpSetOptionInfoV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionID        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	OptionInfo      msdhcpm.DHCP_OPTION
}

func (*r_DhcpSetOptionInfoV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetOptionInfoV5 }

// r_DhcpSetOptionInfoV5Response carries the [out] parameters and return value of R_DhcpSetOptionInfoV5.
type r_DhcpSetOptionInfoV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionInfoV5 calls R_DhcpSetOptionInfoV5 (opnum 15) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionInfoV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionID ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, optionInfo msdhcpm.DHCP_OPTION) (err error) {
	req := &r_DhcpSetOptionInfoV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionID:        optionID,
		ClassName:       className,
		VendorName:      vendorName,
		OptionInfo:      optionInfo,
	}
	var resp r_DhcpSetOptionInfoV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionInfoV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionInfoV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
