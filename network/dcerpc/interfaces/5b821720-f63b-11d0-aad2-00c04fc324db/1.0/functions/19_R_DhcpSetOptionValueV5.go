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

// r_DhcpSetOptionValueV5Request carries the [in] parameters of R_DhcpSetOptionValueV5.
type r_DhcpSetOptionValueV5Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
	OptionId        ndr.DWORD
	ClassName       *ndr.WSTR `ndr:"unique"`
	VendorName      *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO
	OptionValue     msdhcpm.DHCP_OPTION_DATA
}

func (*r_DhcpSetOptionValueV5Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpSetOptionValueV5 }

// r_DhcpSetOptionValueV5Response carries the [out] parameters and return value of R_DhcpSetOptionValueV5.
type r_DhcpSetOptionValueV5Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpSetOptionValueV5 calls R_DhcpSetOptionValueV5 (opnum 19) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpSetOptionValueV5(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD, optionId ndr.DWORD, className *ndr.WSTR, vendorName *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO, optionValue msdhcpm.DHCP_OPTION_DATA) (err error) {
	req := &r_DhcpSetOptionValueV5Request{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
		OptionId:        optionId,
		ClassName:       className,
		VendorName:      vendorName,
		ScopeInfo:       scopeInfo,
		OptionValue:     optionValue,
	}
	var resp r_DhcpSetOptionValueV5Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpSetOptionValueV5: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpSetOptionValueV5 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
