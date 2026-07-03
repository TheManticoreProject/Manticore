package functions

import (
	"fmt"

	PerflibV2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/da5a86c5-12c2-4943-ab30-7f74a813d853/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspcq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pcq"
)

// perflibV2CloseQueryHandleRequest carries the [in] parameters of PerflibV2CloseQueryHandle.
type perflibV2CloseQueryHandleRequest struct {
	PhQuery mspcq.PRPC_HQUERY
}

func (*perflibV2CloseQueryHandleRequest) Opnum() uint16 {
	return PerflibV2.OpnumPerflibV2CloseQueryHandle
}

// perflibV2CloseQueryHandleResponse carries the [out] parameters and return value of PerflibV2CloseQueryHandle.
type perflibV2CloseQueryHandleResponse struct {
	PhQuery mspcq.PRPC_HQUERY
	Status  ndr.DWORD `ndr:"retval"`
}

// PerflibV2CloseQueryHandle calls PerflibV2CloseQueryHandle (opnum 4) ([MS-PCQ] 3.1.4.5). It
// closes the query and releases the server-side state; phQuery is [in,out] and the server
// zeroes the returned handle on success (RPC_HQUERY.IsZero reports this).
func PerflibV2CloseQueryHandle(rpc ndr.Invoker, phQuery mspcq.PRPC_HQUERY) (PhQuery mspcq.PRPC_HQUERY, err error) {
	req := &perflibV2CloseQueryHandleRequest{
		PhQuery: phQuery,
	}
	var resp perflibV2CloseQueryHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("PerflibV2CloseQueryHandle: %w", err)
		return
	}
	PhQuery = resp.PhQuery
	if uint32(resp.Status) != PerflibV2.StatusSuccess {
		err = fmt.Errorf("PerflibV2CloseQueryHandle failed: %s", PerflibV2.StatusString(uint32(resp.Status)))
	}
	return
}
