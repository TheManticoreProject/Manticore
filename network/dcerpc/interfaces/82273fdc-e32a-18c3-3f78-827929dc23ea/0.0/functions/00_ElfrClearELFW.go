package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrClearELFWRequest carries the [in] parameters of ElfrClearELFW.
type elfrClearELFWRequest struct {
	LogHandle      mseven.IELF_HANDLE
	BackupFileName *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
}

func (*elfrClearELFWRequest) Opnum() uint16 { return eventlog.OpnumElfrClearELFW }

// ElfrClearELFW calls ElfrClearELFW (opnum 0) ([MS-EVEN] section 3.1.4).
func ElfrClearELFW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, backupFileName *msdtyp.RPC_UNICODE_STRING) (err error) {
	req := &elfrClearELFWRequest{
		LogHandle:      logHandle,
		BackupFileName: backupFileName,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrClearELFW: %w", err)
		return
	}
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrClearELFW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
