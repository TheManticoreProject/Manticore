package functions

import (
	"fmt"

	WindowsShutdown "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d95afe70-a6d5-4259-822e-2c84da1ddb0d/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rsp"
)

// wsdrInitiateShutdownRequest carries the [in] parameters of WsdrInitiateShutdown.
type wsdrInitiateShutdownRequest struct {
	LpMessage       *msrsp.REG_UNICODE_STRING `ndr:"unique"`
	DwGracePeriod   ndr.DWORD
	DwShutdownFlags ndr.DWORD
	DwReason        ndr.DWORD
	LpClientHint    *msrsp.REG_UNICODE_STRING `ndr:"unique"`
}

func (*wsdrInitiateShutdownRequest) Opnum() uint16 { return WindowsShutdown.OpnumWsdrInitiateShutdown }

// wsdrInitiateShutdownResponse carries the [out] parameters and return value of WsdrInitiateShutdown.
type wsdrInitiateShutdownResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// WsdrInitiateShutdown calls WsdrInitiateShutdown (opnum 0) ([MS-RSP] section 3.3.4.1). It initiates the shutdown of the
// server after dwGracePeriod seconds, per the dwShutdownFlags and dwReason codes.
func WsdrInitiateShutdown(rpc ndr.Invoker, lpMessage *msrsp.REG_UNICODE_STRING, dwGracePeriod ndr.DWORD, dwShutdownFlags ndr.DWORD, dwReason ndr.DWORD, lpClientHint *msrsp.REG_UNICODE_STRING) (err error) {
	req := &wsdrInitiateShutdownRequest{
		LpMessage:       lpMessage,
		DwGracePeriod:   dwGracePeriod,
		DwShutdownFlags: dwShutdownFlags,
		DwReason:        dwReason,
		LpClientHint:    lpClientHint,
	}
	var resp wsdrInitiateShutdownResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WsdrInitiateShutdown: %w", err)
		return
	}
	if uint32(resp.Status) != WindowsShutdown.StatusSuccess {
		err = fmt.Errorf("WsdrInitiateShutdown failed: %s", WindowsShutdown.StatusString(uint32(resp.Status)))
	}
	return
}
