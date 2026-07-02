package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrRegisterEventSourceARequest carries the [in] parameters of ElfrRegisterEventSourceA.
type elfrRegisterEventSourceARequest struct {
	UNCServerName ndr.STR `ndr:"unique"`
	ModuleName    mseven.RPC_STRING
	RegModuleName mseven.RPC_STRING
	MajorVersion  ndr.DWORD
	MinorVersion  ndr.DWORD
}

func (*elfrRegisterEventSourceARequest) Opnum() uint16 { return eventlog.OpnumElfrRegisterEventSourceA }

// ElfrRegisterEventSourceA calls ElfrRegisterEventSourceA (opnum 15) ([MS-EVEN] section 3.1.4).
func ElfrRegisterEventSourceA(rpc ndr.Invoker, uNCServerName ndr.STR, moduleName mseven.RPC_STRING, regModuleName mseven.RPC_STRING, majorVersion ndr.DWORD, minorVersion ndr.DWORD) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrRegisterEventSourceARequest{
		UNCServerName: uNCServerName,
		ModuleName:    moduleName,
		RegModuleName: regModuleName,
		MajorVersion:  majorVersion,
		MinorVersion:  minorVersion,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrRegisterEventSourceA: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrRegisterEventSourceA failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
