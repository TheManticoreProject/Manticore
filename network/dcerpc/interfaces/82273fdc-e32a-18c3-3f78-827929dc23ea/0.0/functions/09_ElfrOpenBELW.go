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

// elfrOpenBELWRequest carries the [in] parameters of ElfrOpenBELW.
type elfrOpenBELWRequest struct {
	UNCServerName  ndr.WSTR `ndr:"unique"`
	BackupFileName msdtyp.RPC_UNICODE_STRING
	MajorVersion   ndr.DWORD
	MinorVersion   ndr.DWORD
}

func (*elfrOpenBELWRequest) Opnum() uint16 { return eventlog.OpnumElfrOpenBELW }

// ElfrOpenBELW calls ElfrOpenBELW (opnum 9) ([MS-EVEN] section 3.1.4).
func ElfrOpenBELW(rpc ndr.Invoker, uNCServerName ndr.WSTR, backupFileName msdtyp.RPC_UNICODE_STRING, majorVersion ndr.DWORD, minorVersion ndr.DWORD) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrOpenBELWRequest{
		UNCServerName:  uNCServerName,
		BackupFileName: backupFileName,
		MajorVersion:   majorVersion,
		MinorVersion:   minorVersion,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrOpenBELW: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrOpenBELW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
