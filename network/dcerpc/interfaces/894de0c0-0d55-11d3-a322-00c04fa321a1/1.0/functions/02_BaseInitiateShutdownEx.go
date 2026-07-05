package functions

import (
	"fmt"

	InitShutdown "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/894de0c0-0d55-11d3-a322-00c04fa321a1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rsp"
)

// baseInitiateShutdownExRequest carries the [in] parameters of BaseInitiateShutdownEx.
type baseInitiateShutdownExRequest struct {
	ServerName           *ndr.WSTR                 `ndr:"unique"`
	LpMessage            *msrsp.REG_UNICODE_STRING `ndr:"unique"`
	DwTimeout            ndr.DWORD
	BForceAppsClosed     uint8
	BRebootAfterShutdown uint8
	DwReason             ndr.DWORD
}

func (*baseInitiateShutdownExRequest) Opnum() uint16 { return InitShutdown.OpnumBaseInitiateShutdownEx }

// baseInitiateShutdownExResponse carries the [out] parameters and return value of BaseInitiateShutdownEx.
type baseInitiateShutdownExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseInitiateShutdownEx calls BaseInitiateShutdownEx (opnum 2) ([MS-RSP] section 3.2.4.3). Like BaseInitiateShutdown but also
// records dwReason, a shutdown reason code (see [MS-RSP] section 2.3).
func BaseInitiateShutdownEx(rpc ndr.Invoker, serverName *ndr.WSTR, lpMessage *msrsp.REG_UNICODE_STRING, dwTimeout ndr.DWORD, bForceAppsClosed uint8, bRebootAfterShutdown uint8, dwReason ndr.DWORD) (err error) {
	req := &baseInitiateShutdownExRequest{
		ServerName:           serverName,
		LpMessage:            lpMessage,
		DwTimeout:            dwTimeout,
		BForceAppsClosed:     bForceAppsClosed,
		BRebootAfterShutdown: bRebootAfterShutdown,
		DwReason:             dwReason,
	}
	var resp baseInitiateShutdownExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseInitiateShutdownEx: %w", err)
		return
	}
	if uint32(resp.Status) != InitShutdown.StatusSuccess {
		err = fmt.Errorf("BaseInitiateShutdownEx failed: %s", InitShutdown.StatusString(uint32(resp.Status)))
	}
	return
}
