package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpRestoreDatabaseRequest carries the [in] parameters of R_DhcpRestoreDatabase.
type r_DhcpRestoreDatabaseRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Path            ndr.WSTR
}

func (*r_DhcpRestoreDatabaseRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpRestoreDatabase }

// r_DhcpRestoreDatabaseResponse carries the [out] parameters and return value of R_DhcpRestoreDatabase.
type r_DhcpRestoreDatabaseResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpRestoreDatabase calls R_DhcpRestoreDatabase (opnum 45) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpRestoreDatabase(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, path ndr.WSTR) (err error) {
	req := &r_DhcpRestoreDatabaseRequest{
		ServerIpAddress: serverIpAddress,
		Path:            path,
	}
	var resp r_DhcpRestoreDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpRestoreDatabase: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpRestoreDatabase failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
