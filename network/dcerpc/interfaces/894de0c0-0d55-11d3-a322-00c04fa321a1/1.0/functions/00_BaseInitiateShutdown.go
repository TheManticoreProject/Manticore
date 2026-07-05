package functions

import (
	"fmt"

	InitShutdown "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/894de0c0-0d55-11d3-a322-00c04fa321a1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rsp"
)

// baseInitiateShutdownRequest carries the [in] parameters of BaseInitiateShutdown.
type baseInitiateShutdownRequest struct {
	ServerName           *ndr.WSTR                 `ndr:"unique"`
	LpMessage            *msrsp.REG_UNICODE_STRING `ndr:"unique"`
	DwTimeout            ndr.DWORD
	BForceAppsClosed     uint8
	BRebootAfterShutdown uint8
}

func (*baseInitiateShutdownRequest) Opnum() uint16 { return InitShutdown.OpnumBaseInitiateShutdown }

// baseInitiateShutdownResponse carries the [out] parameters and return value of BaseInitiateShutdown.
type baseInitiateShutdownResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseInitiateShutdown calls BaseInitiateShutdown (opnum 0) ([MS-RSP] section 3.2.4.1). It initiates the shutdown of the
// server, optionally displaying lpMessage for dwTimeout seconds first.
func BaseInitiateShutdown(rpc ndr.Invoker, serverName *ndr.WSTR, lpMessage *msrsp.REG_UNICODE_STRING, dwTimeout ndr.DWORD, bForceAppsClosed uint8, bRebootAfterShutdown uint8) (err error) {
	req := &baseInitiateShutdownRequest{
		ServerName:           serverName,
		LpMessage:            lpMessage,
		DwTimeout:            dwTimeout,
		BForceAppsClosed:     bForceAppsClosed,
		BRebootAfterShutdown: bRebootAfterShutdown,
	}
	var resp baseInitiateShutdownResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseInitiateShutdown: %w", err)
		return
	}
	if uint32(resp.Status) != InitShutdown.StatusSuccess {
		err = fmt.Errorf("BaseInitiateShutdown failed: %s", InitShutdown.StatusString(uint32(resp.Status)))
	}
	return
}
