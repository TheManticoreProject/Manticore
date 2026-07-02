package functions

import (
	"fmt"

	eventlog "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82273fdc-e32a-18c3-3f78-827929dc23ea/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even"
)

// elfrGetLogInformationRequest carries the [in] parameters of ElfrGetLogInformation.
type elfrGetLogInformationRequest struct {
	LogHandle mseven.IELF_HANDLE
	InfoLevel ndr.DWORD
	CbBufSize ndr.DWORD
}

func (*elfrGetLogInformationRequest) Opnum() uint16 { return eventlog.OpnumElfrGetLogInformation }

// elfrGetLogInformationResponse carries the [out] parameters and return value of ElfrGetLogInformation.
type elfrGetLogInformationResponse struct {
	LpBuffer       []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// ElfrGetLogInformation calls ElfrGetLogInformation (opnum 22) ([MS-EVEN] section 3.1.4).
func ElfrGetLogInformation(rpc ndr.Invoker, logHandle mseven.IELF_HANDLE, infoLevel ndr.DWORD, cbBufSize ndr.DWORD) (LpBuffer []uint8, PcbBytesNeeded ndr.DWORD, err error) {
	req := &elfrGetLogInformationRequest{
		LogHandle: logHandle,
		InfoLevel: infoLevel,
		CbBufSize: cbBufSize,
	}
	var resp elfrGetLogInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ElfrGetLogInformation: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != eventlog.StatusSuccess {
		err = fmt.Errorf("ElfrGetLogInformation failed: %s", eventlog.StatusString(uint32(resp.Status)))
	}
	return
}
