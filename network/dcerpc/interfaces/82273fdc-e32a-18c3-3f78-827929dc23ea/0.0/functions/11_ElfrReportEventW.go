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

// elfrReportEventWRequest carries the [in] parameters of ElfrReportEventW.
type elfrReportEventWRequest struct {
	LogHandle     mseven.IELF_HANDLE
	Time          ndr.DWORD
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
	TimeWritten   *ndr.DWORD `ndr:"unique"`
}

func (*elfrReportEventWRequest) Opnum() uint16 { return eventlog.OpnumElfrReportEventW }

// elfrReportEventWResponse carries the [out] parameters and return value of ElfrReportEventW.
type elfrReportEventWResponse struct {
	RecordNumber *ndr.DWORD `ndr:"unique"`
	TimeWritten  *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// ElfrReportEventW calls ElfrReportEventW (opnum 11) ([MS-EVEN] section 3.1.4).
func ElfrReportEventW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, time ndr.DWORD, eventType uint16, eventCategory uint16, eventID ndr.DWORD, numStrings uint16, dataSize ndr.DWORD, computerName msdtyp.RPC_UNICODE_STRING, userSID *msdtyp.RPC_SID, strings []*msdtyp.RPC_UNICODE_STRING, data []uint8, flags uint16, recordNumber *ndr.DWORD, timeWritten *ndr.DWORD) (RecordNumber *ndr.DWORD, TimeWritten *ndr.DWORD, err error) {
	req := &elfrReportEventWRequest{
		LogHandle:     logHandle,
		Time:          time,
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
		TimeWritten:   timeWritten,
	}
	var resp elfrReportEventWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrReportEventW: %w", err)
		return
	}
	RecordNumber = resp.RecordNumber
	TimeWritten = resp.TimeWritten
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrReportEventW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
