package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrOldestRecordRequest carries the [in] parameters of ElfrOldestRecord.
type elfrOldestRecordRequest struct {
	LogHandle mseven.IELF_HANDLE
}

func (*elfrOldestRecordRequest) Opnum() uint16 { return eventlog.OpnumElfrOldestRecord }

// elfrOldestRecordResponse carries the [out] parameters and return value of ElfrOldestRecord.
type elfrOldestRecordResponse struct {
	OldestRecordNumber ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// ElfrOldestRecord calls ElfrOldestRecord (opnum 5) ([MS-EVEN] section 3.1.4).
func ElfrOldestRecord(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE) (OldestRecordNumber ndr.DWORD, err error) {
	req := &elfrOldestRecordRequest{
		LogHandle: logHandle,
	}
	var resp elfrOldestRecordResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrOldestRecord: %w", err)
		return
	}
	OldestRecordNumber = resp.OldestRecordNumber
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrOldestRecord failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
