package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f874532d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpGetVersionRequest carries the [in] parameters of R_DhcpGetVersion.
type r_DhcpGetVersionRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
}

func (*r_DhcpGetVersionRequest) Opnum() uint16 { return dhcpsrv.OpnumR_DhcpGetVersion }

// r_DhcpGetVersionResponse carries the [out] parameters and return value of R_DhcpGetVersion.
type r_DhcpGetVersionResponse struct {
	MajorVersion ndr.DWORD
	MinorVersion ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// R_DhcpGetVersion calls R_DhcpGetVersion (opnum 28) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpGetVersion(rpc ndr.Invoker, serverIpAddress *ndr.WSTR) (MajorVersion ndr.DWORD, MinorVersion ndr.DWORD, err error) {
	req := &r_DhcpGetVersionRequest{
		ServerIpAddress: serverIpAddress,
	}
	var resp r_DhcpGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpGetVersion: %w", err)
		return
	}
	MajorVersion = resp.MajorVersion
	MinorVersion = resp.MinorVersion
	if uint32(resp.Status) != dhcpsrv.StatusSuccess {
		err = fmt.Errorf("R_DhcpGetVersion failed: %s", dhcpsrv.StatusString(uint32(resp.Status)))
	}
	return
}
