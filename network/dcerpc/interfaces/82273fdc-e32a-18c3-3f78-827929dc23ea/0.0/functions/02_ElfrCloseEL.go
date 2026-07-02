package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrCloseELRequest carries the [in] parameters of ElfrCloseEL.
type elfrCloseELRequest struct {
	LogHandle mseven.IELF_HANDLE
}

func (*elfrCloseELRequest) Opnum() uint16 { return eventlog.OpnumElfrCloseEL }

// ElfrCloseEL calls ElfrCloseEL (opnum 2) ([MS-EVEN] section 3.1.4).
func ElfrCloseEL(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrCloseELRequest{
		LogHandle: logHandle,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrCloseEL: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrCloseEL failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
