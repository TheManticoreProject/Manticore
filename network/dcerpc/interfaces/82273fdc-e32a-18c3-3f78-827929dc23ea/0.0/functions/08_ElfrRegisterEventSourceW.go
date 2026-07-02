package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrRegisterEventSourceWRequest carries the [in] parameters of ElfrRegisterEventSourceW.
type elfrRegisterEventSourceWRequest struct {
	UNCServerName ndr.WSTR `ndr:"unique"`
	ModuleName    dtyp.RPC_UNICODE_STRING
	RegModuleName dtyp.RPC_UNICODE_STRING
	MajorVersion  ndr.DWORD
	MinorVersion  ndr.DWORD
}

func (*elfrRegisterEventSourceWRequest) Opnum() uint16 { return eventlog.OpnumElfrRegisterEventSourceW }

// ElfrRegisterEventSourceW calls ElfrRegisterEventSourceW (opnum 8) ([MS-EVEN] section 3.1.4).
func ElfrRegisterEventSourceW(rpc ndr.Invoker, uNCServerName ndr.WSTR, moduleName dtyp.RPC_UNICODE_STRING, regModuleName dtyp.RPC_UNICODE_STRING, majorVersion ndr.DWORD, minorVersion ndr.DWORD) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrRegisterEventSourceWRequest{
		UNCServerName: uNCServerName,
		ModuleName:    moduleName,
		RegModuleName: regModuleName,
		MajorVersion:  majorVersion,
		MinorVersion:  minorVersion,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrRegisterEventSourceW: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrRegisterEventSourceW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
