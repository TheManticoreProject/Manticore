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

// r_DhcpServerGetConfigV4Request carries the [in] parameters of R_DhcpServerGetConfigV4.
type r_DhcpServerGetConfigV4Request struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpServerGetConfigV4Request) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpServerGetConfigV4 }

// r_DhcpServerGetConfigV4Response carries the [out] parameters and return value of R_DhcpServerGetConfigV4.
type r_DhcpServerGetConfigV4Response struct {
	ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO_V4 `ndr:"unique"`
	Status     ndr.DWORD                           `ndr:"retval"`
}

// R_DhcpServerGetConfigV4 calls R_DhcpServerGetConfigV4 (opnum 40) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpServerGetConfigV4(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (ConfigInfo *msdhcpm.DHCP_SERVER_CONFIG_INFO_V4, err error) {
	req := &r_DhcpServerGetConfigV4Request{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpServerGetConfigV4Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpServerGetConfigV4: %w", err)
		return
	}
	ConfigInfo = resp.ConfigInfo
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpServerGetConfigV4 failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
