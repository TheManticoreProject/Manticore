package functions

import (
	"fmt"

	PerflibV2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/da5a86c5-12c2-4943-ab30-7f74a813d853/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspcq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pcq"
)

// perflibV2QueryCounterInfoRequest carries the [in] parameters of PerflibV2QueryCounterInfo.
type perflibV2QueryCounterInfoRequest struct {
	HQuery   mspcq.RPC_HQUERY
	DwInSize ndr.DWORD
}

func (*perflibV2QueryCounterInfoRequest) Opnum() uint16 {
	return PerflibV2.OpnumPerflibV2QueryCounterInfo
}

// perflibV2QueryCounterInfoResponse carries the [out] parameters and return value of PerflibV2QueryCounterInfo.
type perflibV2QueryCounterInfoResponse struct {
	PdwOutSize ndr.DWORD
	PdwRtnSize ndr.DWORD
	// [out, size_is(dwInSize), length_is(*pdwOutSize)] byte buffer: top-level [ref] pointer
	// to a conformant-varying array whose counts arrive inline (see opnum 0).
	LpData []uint8   `ndr:"ref,varying"`
	Status ndr.DWORD `ndr:"retval"`
}

// PerflibV2QueryCounterInfo calls PerflibV2QueryCounterInfo (opnum 5) ([MS-PCQ] 3.1.4.6).
// A dwInSize of 0 is the size-probe form; the server returns ERROR_NOT_ENOUGH_MEMORY with the
// required buffer size in pdwRtnSize, which this stub tolerates as a non-error.
func PerflibV2QueryCounterInfo(rpc ndr.Invoker, hQuery mspcq.RPC_HQUERY, dwInSize ndr.DWORD) (PdwOutSize ndr.DWORD, PdwRtnSize ndr.DWORD, LpData []uint8, err error) {
	req := &perflibV2QueryCounterInfoRequest{
		HQuery:   hQuery,
		DwInSize: dwInSize,
	}
	var resp perflibV2QueryCounterInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PerflibV2QueryCounterInfo: %w", err)
		return
	}
	PdwOutSize = resp.PdwOutSize
	PdwRtnSize = resp.PdwRtnSize
	LpData = resp.LpData
	if s := uint32(resp.Status); s != PerflibV2.StatusSuccess && s != PerflibV2.ErrorNotEnoughMemory {
		err = fmt.Errorf("PerflibV2QueryCounterInfo failed: %s", PerflibV2.StatusString(s))
	}
	return
}
