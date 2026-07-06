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

// elfrReportEventExWRequest carries the [in] parameters of ElfrReportEventExW.
type elfrReportEventExWRequest struct {
	LogHandle     mseven.IELF_HANDLE
	TimeGenerated msdtyp.FILETIME
	EventType     uint16
	EventCategory uint16
	EventID       ndr.DWORD
	NumStrings    uint16
	DataSize      ndr.DWORD
	ComputerName  msdtyp.RPC_UNICODE_STRING
	UserSID       *msdtyp.RPC_SID              `ndr:"unique"`
	Strings       []*msdtyp.RPC_UNICODE_STRING `ndr:"unique,elem=unique,size_is=NumStrings"`
	Data          []uint8                      `ndr:"unique,size_is=DataSize"`
	Flags         uint16
	RecordNumber  *ndr.DWORD `ndr:"unique"`
}

func (*elfrReportEventExWRequest) Opnum() uint16 { return eventlog.OpnumElfrReportEventExW }

// elfrReportEventExWResponse carries the [out] parameters and return value of ElfrReportEventExW.
type elfrReportEventExWResponse struct {
	RecordNumber *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// ElfrReportEventExW calls ElfrReportEventExW (opnum 25) ([MS-EVEN] section 3.1.4).
func ElfrReportEventExW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, timeGenerated msdtyp.FILETIME, eventType uint16, eventCategory uint16, eventID ndr.DWORD, numStrings uint16, dataSize ndr.DWORD, computerName msdtyp.RPC_UNICODE_STRING, userSID *msdtyp.RPC_SID, strings []*msdtyp.RPC_UNICODE_STRING, data []uint8, flags uint16, recordNumber *ndr.DWORD) (RecordNumber *ndr.DWORD, err error) {
	req := &elfrReportEventExWRequest{
		LogHandle:     logHandle,
		TimeGenerated: timeGenerated,
		EventType:     eventType,
		EventCategory: eventCategory,
		EventID:       eventID,
		NumStrings:    numStrings,
		DataSize:      dataSize,
		ComputerName:  computerName,
		UserSID:       userSID,
		Strings:       strings,
		Data:          data,
		Flags:         flags,
		RecordNumber:  recordNumber,
	}
	var resp elfrReportEventExWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrReportEventExW: %w", err)
		return
	}
	RecordNumber = resp.RecordNumber
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrReportEventExW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
