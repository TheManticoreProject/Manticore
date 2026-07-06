package functions

// IDL source: [MS-PCQ] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-pcq/dcee10e3-0512-495e-9566-26e56cc21c5c
// A fetched copy is kept at ms-pcq.idl in the interface directory.

import (
	"fmt"

	PerflibV2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/da5a86c5-12c2-4943-ab30-7f74a813d853/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspcq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pcq"
)

// perflibV2ValidateCountersRequest carries the [in] parameters of PerflibV2ValidateCounters.
type perflibV2ValidateCountersRequest struct {
	HQuery   mspcq.RPC_HQUERY
	DwInSize ndr.DWORD
	LpData   []uint8 `ndr:"ref,size_is=DwInSize"`
	DwAdd    ndr.DWORD
}

func (*perflibV2ValidateCountersRequest) Opnum() uint16 {
	return PerflibV2.OpnumPerflibV2ValidateCounters
}

// perflibV2ValidateCountersResponse carries the [out] parameters and return value of PerflibV2ValidateCounters.
type perflibV2ValidateCountersResponse struct {
	// [in, out, size_is(dwInSize)] byte buffer: top-level [ref] pointer to a conformant
	// (non-varying) array. maximum_count (=dwInSize) arrives inline on the response, so the
	// decoder reads it directly; the size_is hint on the request tag drives marshalling only.
	LpData []uint8   `ndr:"ref,conformant"`
	Status ndr.DWORD `ndr:"retval"`
}

// PerflibV2ValidateCounters calls PerflibV2ValidateCounters (opnum 7) ([MS-PCQ] 3.1.4.8). lpData
// is an in/out buffer of exactly dwInSize bytes; pass len(lpData) == dwInSize.
func PerflibV2ValidateCounters(rpc ndr.Invoker, hQuery mspcq.RPC_HQUERY, dwInSize ndr.DWORD, lpData []uint8, dwAdd ndr.DWORD) (LpData []uint8, err error) {
	req := &perflibV2ValidateCountersRequest{
		HQuery:   hQuery,
		DwInSize: dwInSize,
		LpData:   lpData,
		DwAdd:    dwAdd,
	}
	var resp perflibV2ValidateCountersResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PerflibV2ValidateCounters: %w", err)
		return
	}
	LpData = resp.LpData
	if uint32(resp.Status) != PerflibV2.StatusSuccess {
		err = fmt.Errorf("PerflibV2ValidateCounters failed: %s", PerflibV2.StatusString(uint32(resp.Status)))
	}
	return
}
