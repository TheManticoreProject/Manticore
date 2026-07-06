package functions

// IDL source: [MS-EVEN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even/0d0bee9c-dac5-46d9-b19b-2087826c02db
// A fetched copy is kept at ms-even.idl in the interface directory.

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrBackupELFWRequest carries the [in] parameters of ElfrBackupELFW.
type elfrBackupELFWRequest struct {
	LogHandle      mseven.IELF_HANDLE
	BackupFileName msdtyp.RPC_UNICODE_STRING
}

func (*elfrBackupELFWRequest) Opnum() uint16 { return eventlog.OpnumElfrBackupELFW }

// ElfrBackupELFW calls ElfrBackupELFW (opnum 1) ([MS-EVEN] section 3.1.4).
func ElfrBackupELFW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, backupFileName msdtyp.RPC_UNICODE_STRING) (err error) {
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
