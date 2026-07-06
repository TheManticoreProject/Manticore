package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrOpenELWRequest carries the [in] parameters of ElfrOpenELW.
type elfrOpenELWRequest struct {
	UNCServerName ndr.WSTR `ndr:"unique"`
	ModuleName    msdtyp.RPC_UNICODE_STRING
	RegModuleName msdtyp.RPC_UNICODE_STRING
	MajorVersion  ndr.DWORD
	MinorVersion  ndr.DWORD
}

func (*elfrOpenELWRequest) Opnum() uint16 { return eventlog.OpnumElfrOpenELW }

// ElfrOpenELW calls ElfrOpenELW (opnum 7) ([MS-EVEN] section 3.1.4).
func ElfrOpenELW(rpc ndr.Invoker, uNCServerName ndr.WSTR, moduleName msdtyp.RPC_UNICODE_STRING, regModuleName msdtyp.RPC_UNICODE_STRING, majorVersion ndr.DWORD, minorVersion ndr.DWORD) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrOpenELWRequest{
		UNCServerName: uNCServerName,
		ModuleName:    moduleName,
		RegModuleName: regModuleName,
		MajorVersion:  majorVersion,
		MinorVersion:  minorVersion,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrOpenELW: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrOpenELW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
