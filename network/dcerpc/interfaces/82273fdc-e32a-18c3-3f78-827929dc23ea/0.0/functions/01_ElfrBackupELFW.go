package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrBackupELFWRequest carries the [in] parameters of ElfrBackupELFW.
type elfrBackupELFWRequest struct {
	LogHandle      mseven.IELF_HANDLE
	BackupFileName dtyp.RPC_UNICODE_STRING
}

func (*elfrBackupELFWRequest) Opnum() uint16 { return eventlog.OpnumElfrBackupELFW }

// ElfrBackupELFW calls ElfrBackupELFW (opnum 1) ([MS-EVEN] section 3.1.4).
func ElfrBackupELFW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, backupFileName dtyp.RPC_UNICODE_STRING) (err error) {
	req := &elfrBackupELFWRequest{
		LogHandle:      logHandle,
		BackupFileName: backupFileName,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrBackupELFW: %w", err)
		return
	}
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrBackupELFW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
