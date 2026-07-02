package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiBackupClusterDatabaseRequest carries the [in] parameters of ApiBackupClusterDatabase.
type apiBackupClusterDatabaseRequest struct {
	LpszPathName ndr.WSTR
}

func (*apiBackupClusterDatabaseRequest) Opnum() uint16 { return clusapi.OpnumApiBackupClusterDatabase }

// apiBackupClusterDatabaseResponse carries the [out] parameters and return value of ApiBackupClusterDatabase.
type apiBackupClusterDatabaseResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiBackupClusterDatabase calls ApiBackupClusterDatabase (opnum 104) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiBackupClusterDatabase(rpc ndr.Invoker, lpszPathName ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiBackupClusterDatabaseRequest{
		LpszPathName: lpszPathName,
	}
	var resp apiBackupClusterDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiBackupClusterDatabase: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiBackupClusterDatabase failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
