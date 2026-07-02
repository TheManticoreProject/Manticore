package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrReadELWRequest carries the [in] parameters of ElfrReadELW.
type elfrReadELWRequest struct {
	LogHandle           mseven.IELF_HANDLE
	ReadFlags           ndr.DWORD
	RecordOffset        ndr.DWORD
	NumberOfBytesToRead mseven.RULONG
}

func (*elfrReadELWRequest) Opnum() uint16 { return eventlog.OpnumElfrReadELW }

// elfrReadELWResponse carries the [out] parameters and return value of ElfrReadELW.
type elfrReadELWResponse struct {
	Buffer                 []uint8 `ndr:"ref,size_is=NumberOfBytesToRead"`
	NumberOfBytesRead      ndr.DWORD
	MinNumberOfBytesNeeded ndr.DWORD
	Status                 ndr.DWORD `ndr:"retval"`
}

// ElfrReadELW calls ElfrReadELW (opnum 10) ([MS-EVEN] section 3.1.4).
func ElfrReadELW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, readFlags ndr.DWORD, recordOffset ndr.DWORD, numberOfBytesToRead mseven.RULONG) (Buffer []uint8, NumberOfBytesRead ndr.DWORD, MinNumberOfBytesNeeded ndr.DWORD, err error) {
	req := &elfrReadELWRequest{
		LogHandle:           logHandle,
		ReadFlags:           readFlags,
		RecordOffset:        recordOffset,
		NumberOfBytesToRead: numberOfBytesToRead,
	}
	var resp elfrReadELWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrReadELW: %w", err)
		return
	}
	Buffer = resp.Buffer
	NumberOfBytesRead = resp.NumberOfBytesRead
	MinNumberOfBytesNeeded = resp.MinNumberOfBytesNeeded
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrReadELW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
