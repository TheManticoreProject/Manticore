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

// r_DhcpSetOptionValuesV5Request carries the [in] parameters of R_DhcpSetOptionValuesV5.
type r_DhcpSetOptionValuesV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
	OptionValues    msdhcpm.DHCP_OPTION_VALUE_ARRAY
}

func (*r_DhcpSetOptionValuesV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetOptionValuesV5 }

// r_DhcpSetOptionValuesV5Response carries the [out] parameters and return value of R_DhcpSetOptionValuesV5.
type r_DhcpSetOptionValuesV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionValuesV5 calls R_DhcpSetOptionValuesV5 (opnum 20) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionValuesV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO, optionValues msdhcpm.DHCP_OPTION_VALUE_ARRAY) (err error) {
	req := &r_DhcpSetOptionValuesV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		ClassName:       className,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
		OptionValues:    optionValues,
	}
	var resp r_DhcpSetOptionValuesV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionValuesV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionValuesV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
