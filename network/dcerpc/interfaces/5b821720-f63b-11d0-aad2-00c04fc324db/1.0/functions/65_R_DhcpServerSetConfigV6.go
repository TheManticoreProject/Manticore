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

// r_DhcpServerSetConfigV6Request carries the [in] parameters of R_DhcpServerSetConfigV6.
type r_DhcpServerSetConfigV6Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	ScopeInfo       msdhcpm.DHCP_OPTION_SCOPE_INFO6
	FieldsToSet     ndr.DWORD
	ConfigInfo      msdhcpm.DHCP_SERVER_CONFIG_INFO_V6
}

func (*r_DhcpServerSetConfigV6Request) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpServerSetConfigV6 }

// r_DhcpServerSetConfigV6Response carries the [out] parameters and return value of R_DhcpServerSetConfigV6.
type r_DhcpServerSetConfigV6Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpServerSetConfigV6 calls R_DhcpServerSetConfigV6 (opnum 65) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerSetConfigV6(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, scopeInfo msdhcpm.DHCP_OPTION_SCOPE_INFO6, fieldsToSet ndr.DWORD, configInfo msdhcpm.DHCP_SERVER_CONFIG_INFO_V6) (err error) {
	req := &r_DhcpServerSetConfigV6Request{
		ServerIpAddress: serverIpAddress,
		ScopeInfo:       scopeInfo,
		FieldsToSet:     fieldsToSet,
		ConfigInfo:      configInfo,
	}
	var resp r_DhcpServerSetConfigV6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerSetConfigV6: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerSetConfigV6 failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
