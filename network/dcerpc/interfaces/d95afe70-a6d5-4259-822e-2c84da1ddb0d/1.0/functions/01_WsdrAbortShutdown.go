package functions

// IDL source: [MS-RSP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rsp/012b9c19-5a6f-4d4f-8dd1-a344123a3337
// A fetched copy is kept at ms-rsp.idl in the interface directory.

import (
	"fmt"

	WindowsShutdown "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d95afe70-a6d5-4259-822e-2c84da1ddb0d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rsp"
)

// wsdrAbortShutdownRequest carries the [in] parameters of WsdrAbortShutdown.
type wsdrAbortShutdownRequest struct {
	LpClientHint *msrsp.REG_UNICODE_STRING `ndr:"unique"`
}

func (*wsdrAbortShutdownRequest) Opnum() uint16 { return WindowsShutdown.OpnumWsdrAbortShutdown }

// wsdrAbortShutdownResponse carries the [out] parameters and return value of WsdrAbortShutdown.
type wsdrAbortShutdownResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// WsdrAbortShutdown calls WsdrAbortShutdown (opnum 1) ([MS-RSP] section 3.3.4.2). It stops a system shutdown that was
// previously requested with WsdrInitiateShutdown.
func WsdrAbortShutdown(rpc ndr.Invoker, lpClientHint *msrsp.REG_UNICODE_STRING) (err error) {
	req := &wsdrAbortShutdownRequest{
		LpClientHint: lpClientHint,
	}
	var resp wsdrAbortShutdownResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WsdrAbortShutdown: %w", err)
		return
	}
	if uint32(resp.Status) != WindowsShutdown.StatusSuccess {
		err = fmt.Errorf("WsdrAbortShutdown failed: %s", WindowsShutdown.StatusString(uint32(resp.Status)))
	}
	return
}
