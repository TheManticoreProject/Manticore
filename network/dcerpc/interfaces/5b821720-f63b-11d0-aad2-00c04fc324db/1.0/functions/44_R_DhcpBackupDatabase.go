package functions

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_DhcpBackupDatabaseRequest carries the [in] parameters of R_DhcpBackupDatabase.
type r_DhcpBackupDatabaseRequest struct {
	ServerIpAddress *ndr.WSTR `ndr:"unique"`
	Path            ndr.WSTR
}

func (*r_DhcpBackupDatabaseRequest) Opnum() uint16 { return dhcpsrv2.OpnumR_DhcpBackupDatabase }

// r_DhcpBackupDatabaseResponse carries the [out] parameters and return value of R_DhcpBackupDatabase.
type r_DhcpBackupDatabaseResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_DhcpBackupDatabase calls R_DhcpBackupDatabase (opnum 44) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpBackupDatabase(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, path ndr.WSTR) (err error) {
	req := &r_DhcpBackupDatabaseRequest{
		ServerIpAddress: serverIpAddress,
		Path:            path,
	}
	var resp r_DhcpBackupDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpBackupDatabase: %w", err)
		return
	}
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess {
		err = fmt.Errorf("R_DhcpBackupDatabase failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
