package functions

// IDL source: [MS-EVEN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even/0d0bee9c-dac5-46d9-b19b-2087826c02db
// A fetched copy is kept at ms-even.idl in the interface directory.

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrDeregisterEventSourceRequest carries the [in] parameters of ElfrDeregisterEventSource.
type elfrDeregisterEventSourceRequest struct {
	LogHandle mseven.IELF_HANDLE
}

func (*elfrDeregisterEventSourceRequest) Opnum() uint16 {
	return eventlog.OpnumElfrDeregisterEventSource
}

// ElfrDeregisterEventSource calls ElfrDeregisterEventSource (opnum 3) ([MS-EVEN] section 3.1.4).
func ElfrDeregisterEventSource(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrDeregisterEventSourceRequest{
		LogHandle: logHandle,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrDeregisterEventSource: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrDeregisterEventSource failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
