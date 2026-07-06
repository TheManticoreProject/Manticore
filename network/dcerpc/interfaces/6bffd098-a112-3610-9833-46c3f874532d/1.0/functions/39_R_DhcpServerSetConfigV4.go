package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpServerSetConfigV4Request carries the [in] parameters of R_DhcpServerSetConfigV4.
type r_DhcpServerSetConfigV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	FieldsToSet     ndr.DWORD
	ConfigInfo      msdhcpm.DHCP_SERVER_CONFIG_INFO_V4
}

func (*r_DhcpServerSetConfigV4Request) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpServerSetConfigV4 }

// r_DhcpServerSetConfigV4Response carries the [out] parameters and return value of R_DhcpServerSetConfigV4.
type r_DhcpServerSetConfigV4Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpServerSetConfigV4 calls R_DhcpServerSetConfigV4 (opnum 39) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerSetConfigV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, fieldsToSet ndr.DWORD, configInfo msdhcpm.DHCP_SERVER_CONFIG_INFO_V4) (err error) {
	req := &r_DhcpServerSetConfigV4Request{
		ServerIpAddress: serverIpAddress,
		FieldsToSet:     fieldsToSet,
		ConfigInfo:      configInfo,
	}
	var resp r_DhcpServerSetConfigV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerSetConfigV4: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerSetConfigV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
