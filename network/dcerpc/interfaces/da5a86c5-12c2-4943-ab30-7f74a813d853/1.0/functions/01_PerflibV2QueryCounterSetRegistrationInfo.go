package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	PerflibV2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/da5a86c5-12c2-4943-ab30-7f74a813d853/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// perflibV2QueryCounterSetRegistrationInfoRequest carries the [in] parameters of PerflibV2QueryCounterSetRegistrationInfo.
type perflibV2QueryCounterSetRegistrationInfoRequest struct {
	SzMachine      ndr.WSTR
	CounterSetGuid dtyp.GUID
	RequestCode    ndr.DWORD
	RequestLCID    ndr.DWORD
	DwInSize       ndr.DWORD
}

func (*perflibV2QueryCounterSetRegistrationInfoRequest) Opnum() uint16 {
	return PerflibV2.OpnumPerflibV2QueryCounterSetRegistrationInfo
}

// perflibV2QueryCounterSetRegistrationInfoResponse carries the [out] parameters and return value of PerflibV2QueryCounterSetRegistrationInfo.
type perflibV2QueryCounterSetRegistrationInfoResponse struct {
	PdwOutSize ndr.DWORD
	PdwRtnSize ndr.DWORD
	// [out, size_is(dwInSize), length_is(*pdwOutSize)] byte buffer: top-level [ref] pointer
	// to a conformant-varying array whose counts arrive inline (see opnum 0 for the rationale
	// on omitting size_is/length_is from the response tag).
	LpData []uint8   `ndr:"ref,varying"`
	Status ndr.DWORD `ndr:"retval"`
}

// PerflibV2QueryCounterSetRegistrationInfo calls PerflibV2QueryCounterSetRegistrationInfo (opnum 1) ([MS-PCQ] 3.1.4.2).
// A dwInSize of 0 is the size-probe form; the server returns ERROR_NOT_ENOUGH_MEMORY with the
// required buffer size in pdwRtnSize, which this stub tolerates as a non-error.
func PerflibV2QueryCounterSetRegistrationInfo(rpc ndr.Invoker, szMachine ndr.WSTR, counterSetGuid dtyp.GUID, requestCode ndr.DWORD, requestLCID ndr.DWORD, dwInSize ndr.DWORD) (PdwOutSize ndr.DWORD, PdwRtnSize ndr.DWORD, LpData []uint8, err error) {
	req := &perflibV2QueryCounterSetRegistrationInfoRequest{
		SzMachine:      szMachine,
		CounterSetGuid: counterSetGuid,
		RequestCode:    requestCode,
		RequestLCID:    requestLCID,
		DwInSize:       dwInSize,
	}
	var resp perflibV2QueryCounterSetRegistrationInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PerflibV2QueryCounterSetRegistrationInfo: %w", err)
		return
	}
	PdwOutSize = resp.PdwOutSize
	PdwRtnSize = resp.PdwRtnSize
	LpData = resp.LpData
	if s := uint32(resp.Status); s != PerflibV2.StatusSuccess && s != PerflibV2.ErrorNotEnoughMemory {
		err = fmt.Errorf("PerflibV2QueryCounterSetRegistrationInfo failed: %s", PerflibV2.StatusString(s))
	}
	return
}
