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

// elfrOpenBELARequest carries the [in] parameters of ElfrOpenBELA.
type elfrOpenBELARequest struct {
	UNCServerName  ndr.STR `ndr:"unique"`
	BackupFileName mseven.RPC_STRING
	MajorVersion   ndr.DWORD
	MinorVersion   ndr.DWORD
}

func (*elfrOpenBELARequest) Opnum() uint16 { return eventlog.OpnumElfrOpenBELA }

// ElfrOpenBELA calls ElfrOpenBELA (opnum 16) ([MS-EVEN] section 3.1.4).
func ElfrOpenBELA(rpc ndr.Invoker, uNCServerName ndr.STR, backupFileName mseven.RPC_STRING, majorVersion ndr.DWORD, minorVersion ndr.DWORD) (LogHandle mseven.IELF_HANDLE, err error) {
	req := &elfrOpenBELARequest{
		UNCServerName:  uNCServerName,
		BackupFileName: backupFileName,
		MajorVersion:   majorVersion,
		MinorVersion:   minorVersion,
	}
	var resp handleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrOpenBELA: %w", err)
		return
	}
	LogHandle = resp.LogHandle
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrOpenBELA failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
