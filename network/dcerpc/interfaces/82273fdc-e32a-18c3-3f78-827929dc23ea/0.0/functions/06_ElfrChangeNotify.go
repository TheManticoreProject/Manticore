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

// elfrChangeNotifyRequest carries the [in] parameters of ElfrChangeNotify.
type elfrChangeNotifyRequest struct {
	LogHandle mseven.IELF_HANDLE
	ClientId  mseven.RPC_CLIENT_ID
	Event     ndr.DWORD
}

func (*elfrChangeNotifyRequest) Opnum() uint16 { return eventlog.OpnumElfrChangeNotify }

// ElfrChangeNotify calls ElfrChangeNotify (opnum 6) ([MS-EVEN] section 3.1.4).
func ElfrChangeNotify(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, clientId mseven.RPC_CLIENT_ID, event ndr.DWORD) (err error) {
	req := &elfrChangeNotifyRequest{
		LogHandle: logHandle,
		ClientId:  clientId,
		Event:     event,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrChangeNotify: %w", err)
		return
	}
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrChangeNotify failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
