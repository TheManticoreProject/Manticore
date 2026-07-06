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

// r_DhcpGetAllOptionsRequest carries the [in] parameters of R_DhcpGetAllOptions.
type r_DhcpGetAllOptionsRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Flags           ndr.DWORD
}

func (*r_DhcpGetAllOptionsRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpGetAllOptions }

// r_DhcpGetAllOptionsResponse carries the [out] parameters and return value of R_DhcpGetAllOptions.
type r_DhcpGetAllOptionsResponse struct {
	OptionStruct *msdhcpm.DHCP_ALL_OPTIONS `ndr:"unique"`
	Status       ndr.DWORD                 `ndr:"retval"`
}

// R_DhcpGetAllOptions calls R_DhcpGetAllOptions (opnum 29) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetAllOptions(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, flags ndr.DWORD) (OptionStruct *msdhcpm.DHCP_ALL_OPTIONS, err error) {
	req := &r_DhcpGetAllOptionsRequest{
		ServerIpAddress: serverIpAddress,
		Flags:           flags,
	}
	var resp r_DhcpGetAllOptionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetAllOptions: %w", err)
		return
	}
	OptionStruct = resp.OptionStruct
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetAllOptions failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
