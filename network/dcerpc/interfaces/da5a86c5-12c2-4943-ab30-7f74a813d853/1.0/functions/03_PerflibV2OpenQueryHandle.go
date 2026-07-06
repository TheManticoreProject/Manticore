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

// perflibV2OpenQueryHandleRequest carries the [in] parameters of PerflibV2OpenQueryHandle.
type perflibV2OpenQueryHandleRequest struct {
	SzMachine ndr.WSTR
}

func (*perflibV2OpenQueryHandleRequest) Opnum() uint16 {
	return PerflibV2.OpnumPerflibV2OpenQueryHandle
}

// perflibV2OpenQueryHandleResponse carries the [out] parameters and return value of PerflibV2OpenQueryHandle.
type perflibV2OpenQueryHandleResponse struct {
	PhQuery mspcq.PRPC_HQUERY
	Status  ndr.DWORD `ndr:"retval"`
}

// PerflibV2OpenQueryHandle calls PerflibV2OpenQueryHandle (opnum 3) ([MS-PCQ] 3.1.4.4). It
// creates a query on the server and returns the RPC context handle used by the subsequent
// PerflibV2QueryCounter* calls; release it with PerflibV2CloseQueryHandle.
func PerflibV2OpenQueryHandle(rpc ndr.Invoker, szMachine ndr.WSTR) (PhQuery mspcq.PRPC_HQUERY, err error) {
	req := &perflibV2OpenQueryHandleRequest{
		SzMachine: szMachine,
	}
	var resp perflibV2OpenQueryHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PerflibV2OpenQueryHandle: %w", err)
		return
	}
	PhQuery = resp.PhQuery
	if uint32(resp.Status) != PerflibV2.StatusSuccess {
		err = fmt.Errorf("PerflibV2OpenQueryHandle failed: %s", PerflibV2.StatusString(uint32(resp.Status)))
	}
	return
}
