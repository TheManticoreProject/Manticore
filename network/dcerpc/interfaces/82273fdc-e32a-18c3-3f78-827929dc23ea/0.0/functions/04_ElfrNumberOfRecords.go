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

// elfrNumberOfRecordsRequest carries the [in] parameters of ElfrNumberOfRecords.
type elfrNumberOfRecordsRequest struct {
	LogHandle mseven.IELF_HANDLE
}

func (*elfrNumberOfRecordsRequest) Opnum() uint16 { return eventlog.OpnumElfrNumberOfRecords }

// elfrNumberOfRecordsResponse carries the [out] parameters and return value of ElfrNumberOfRecords.
type elfrNumberOfRecordsResponse struct {
	NumberOfRecords ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ElfrNumberOfRecords calls ElfrNumberOfRecords (opnum 4) ([MS-EVEN] section 3.1.4).
func ElfrNumberOfRecords(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE) (NumberOfRecords ndr.DWORD, err error) {
	req := &elfrNumberOfRecordsRequest{
		LogHandle: logHandle,
	}
	var resp elfrNumberOfRecordsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrNumberOfRecords: %w", err)
		return
	}
	NumberOfRecords = resp.NumberOfRecords
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrNumberOfRecords failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
