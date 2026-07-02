package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrReportEventExARequest carries the [in] parameters of ElfrReportEventExA.
type elfrReportEventExARequest struct {
	LogHandle     mseven.IELF_HANDLE
	TimeGenerated dtyp.FILETIME
	EventType     uint16
	EventCategory uint16
	EventID       ndr.DWORD
	NumStrings    uint16
	DataSize      ndr.DWORD
	ComputerName  mseven.RPC_STRING
	UserSID       *dtyp.RPC_SID        `ndr:"unique"`
	Strings       []*mseven.RPC_STRING `ndr:"unique,elem=unique,size_is=NumStrings"`
	Data          []uint8              `ndr:"unique,size_is=DataSize"`
	Flags         uint16
	RecordNumber  *ndr.DWORD `ndr:"unique"`
}

func (*elfrReportEventExARequest) Opnum() uint16 { return eventlog.OpnumElfrReportEventExA }

// elfrReportEventExAResponse carries the [out] parameters and return value of ElfrReportEventExA.
type elfrReportEventExAResponse struct {
	RecordNumber *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// ElfrReportEventExA calls ElfrReportEventExA (opnum 26) ([MS-EVEN] section 3.1.4).
func ElfrReportEventExA(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, timeGenerated dtyp.FILETIME, eventType uint16, eventCategory uint16, eventID ndr.DWORD, numStrings uint16, dataSize ndr.DWORD, computerName mseven.RPC_STRING, userSID *dtyp.RPC_SID, strings []*mseven.RPC_STRING, data []uint8, flags uint16, recordNumber *ndr.DWORD) (RecordNumber *ndr.DWORD, err error) {
	req := &elfrReportEventExARequest{
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
	var resp elfrReportEventExAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrReportEventExA: %w", err)
		return
	}
	RecordNumber = resp.RecordNumber
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrReportEventExA failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
