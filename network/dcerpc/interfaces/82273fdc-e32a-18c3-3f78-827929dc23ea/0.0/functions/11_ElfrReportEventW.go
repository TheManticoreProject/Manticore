package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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
	ComputerName  dtyp.RPC_UNICODE_STRING
	UserSID       *dtyp.RPC_SID              `ndr:"unique"`
	Strings       []*dtyp.RPC_UNICODE_STRING `ndr:"unique,elem=unique,size_is=NumStrings"`
	Data          []uint8                    `ndr:"unique,size_is=DataSize"`
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
func ElfrReportEventW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, time ndr.DWORD, eventType uint16, eventCategory uint16, eventID ndr.DWORD, numStrings uint16, dataSize ndr.DWORD, computerName dtyp.RPC_UNICODE_STRING, userSID *dtyp.RPC_SID, strings []*dtyp.RPC_UNICODE_STRING, data []uint8, flags uint16, recordNumber *ndr.DWORD, timeWritten *ndr.DWORD) (RecordNumber *ndr.DWORD, TimeWritten *ndr.DWORD, err error) {
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
