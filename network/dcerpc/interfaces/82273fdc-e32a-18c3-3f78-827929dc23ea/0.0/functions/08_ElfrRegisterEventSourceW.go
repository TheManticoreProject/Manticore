package functions

// IDL source: [MS-EVEN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even/0d0bee9c-dac5-46d9-b19b-2087826c02db
// A fetched copy is kept at ms-even.idl in the interface directory.

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrRegisterEventSourceWRequest carries the [in] parameters of ElfrRegisterEventSourceW.
type elfrRegisterEventSourceWRequest struct {
	UNCServerName ndr.WSTR `ndr:"unique"`
	ModuleName    msdtyp.RPC_UNICODE_STRING
	RegModuleName msdtyp.RPC_UNICODE_STRING
	MajorVersion  ndr.DWORD
	MinorVersion  ndr.DWORD
}

func (*elfrRegisterEventSourceWRequest) Opnum() uint16 { return eventlog.OpnumElfrRegisterEventSourceW }

// ElfrRegisterEventSourceW calls ElfrRegisterEventSourceW (opnum 8) ([MS-EVEN] section 3.1.4).
func ElfrRegisterEventSourceW(rpc ndr.Invoker, uNCServerName ndr.WSTR, moduleName msdtyp.RPC_UNICODE_STRING, regModuleName msdtyp.RPC_UNICODE_STRING, majorVersion ndr.DWORD, minorVersion ndr.DWORD) (LogHandle mseven.IELF_HANDLE, err error) {
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
