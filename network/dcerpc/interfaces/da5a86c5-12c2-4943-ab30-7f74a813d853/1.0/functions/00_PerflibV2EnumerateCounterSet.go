package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	PerflibV2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/da5a86c5-12c2-4943-ab30-7f74a813d853/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// perflibV2EnumerateCounterSetRequest carries the [in] parameters of PerflibV2EnumerateCounterSet.
type perflibV2EnumerateCounterSetRequest struct {
	SzMachine ndr.WSTR
	DwInSize  ndr.DWORD
}

func (*perflibV2EnumerateCounterSetRequest) Opnum() uint16 {
	return PerflibV2.OpnumPerflibV2EnumerateCounterSet
}

// perflibV2EnumerateCounterSetResponse carries the [out] parameters and return value of PerflibV2EnumerateCounterSet.
type perflibV2EnumerateCounterSetResponse struct {
	PdwOutSize ndr.DWORD
	PdwRtnSize ndr.DWORD
	// [out, size_is(dwInSize), length_is(*pdwOutSize)] GUID *lpData: a top-level [ref]
	// pointer (an unattributed top-level parameter pointer defaults to [ref]) to a
	// conformant-varying array. maximum_count (=dwInSize) and actual_count (=*pdwOutSize)
	// arrive inline on the wire, so the decoder reads both directly; size_is/length_is are
	// marshal-only hints, and size_is names dwInSize, which is [in]-only and thus absent
	// from this response struct — so the tag is dropped here.
	LpData []dtyp.GUID `ndr:"ref,varying"`
	Status ndr.DWORD   `ndr:"retval"`
}

// PerflibV2EnumerateCounterSet calls PerflibV2EnumerateCounterSet (opnum 0) ([MS-PCQ] 3.1.4.1).
// A dwInSize of 0 is the size-probe form: the server returns ERROR_NOT_ENOUGH_MEMORY with
// the required buffer size in pdwRtnSize, which this stub tolerates as a non-error.
func PerflibV2EnumerateCounterSet(rpc ndr.Invoker, szMachine ndr.WSTR, dwInSize ndr.DWORD) (PdwOutSize ndr.DWORD, PdwRtnSize ndr.DWORD, LpData []dtyp.GUID, err error) {
	req := &perflibV2EnumerateCounterSetRequest{
		SzMachine: szMachine,
		DwInSize:  dwInSize,
	}
	var resp perflibV2EnumerateCounterSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PerflibV2EnumerateCounterSet: %w", err)
		return
	}
	PdwOutSize = resp.PdwOutSize
	PdwRtnSize = resp.PdwRtnSize
	LpData = resp.LpData
	if s := uint32(resp.Status); s != PerflibV2.StatusSuccess && s != PerflibV2.ErrorNotEnoughMemory {
		err = fmt.Errorf("PerflibV2EnumerateCounterSet failed: %s", PerflibV2.StatusString(s))
	}
	return
}
