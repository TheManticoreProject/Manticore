package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrReportEventAndSourceWRequest carries the [in] parameters of ElfrReportEventAndSourceW.
type elfrReportEventAndSourceWRequest struct {
	LogHandle     mseven.IELF_HANDLE
	Time          ndr.DWORD
	EventType     uint16
	EventCategory uint16
	EventID       ndr.DWORD
	SourceName    dtyp.RPC_UNICODE_STRING
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

func (*elfrReportEventAndSourceWRequest) Opnum() uint16 {
	return eventlog.OpnumElfrReportEventAndSourceW
}

// elfrReportEventAndSourceWResponse carries the [out] parameters and return value of ElfrReportEventAndSourceW.
type elfrReportEventAndSourceWResponse struct {
	RecordNumber *ndr.DWORD `ndr:"unique"`
	TimeWritten  *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// ElfrReportEventAndSourceW calls ElfrReportEventAndSourceW (opnum 24) ([MS-EVEN] section 3.1.4).
func ElfrReportEventAndSourceW(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, time ndr.DWORD, eventType uint16, eventCategory uint16, eventID ndr.DWORD, sourceName dtyp.RPC_UNICODE_STRING, numStrings uint16, dataSize ndr.DWORD, computerName dtyp.RPC_UNICODE_STRING, userSID *dtyp.RPC_SID, strings []*dtyp.RPC_UNICODE_STRING, data []uint8, flags uint16, recordNumber *ndr.DWORD, timeWritten *ndr.DWORD) (RecordNumber *ndr.DWORD, TimeWritten *ndr.DWORD, err error) {
	req := &elfrReportEventAndSourceWRequest{
		LogHandle:     logHandle,
		Time:          time,
		EventType:     eventType,
		EventCategory: eventCategory,
		EventID:       eventID,
		SourceName:    sourceName,
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
	var resp elfrReportEventAndSourceWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrReportEventAndSourceW: %w", err)
		return
	}
	RecordNumber = resp.RecordNumber
	TimeWritten = resp.TimeWritten
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrReportEventAndSourceW failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
