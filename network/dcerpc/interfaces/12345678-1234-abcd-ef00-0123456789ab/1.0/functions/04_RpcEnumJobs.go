package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcEnumJobsRequest carries the [in] parameters of RpcEnumJobs.
type rpcEnumJobsRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	FirstJob ndr.DWORD
	NoJobs   ndr.DWORD
	Level    ndr.DWORD
	PJob     []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcEnumJobsRequest) Opnum() uint16 { return winspool.OpnumRpcEnumJobs }

// rpcEnumJobsResponse carries the [out] parameters and return value of RpcEnumJobs.
type rpcEnumJobsResponse struct {
	PJob       []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcEnumJobs calls RpcEnumJobs (opnum 4) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumJobs(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, firstJob ndr.DWORD, noJobs ndr.DWORD, level ndr.DWORD, pJob []uint8, cbBuf ndr.DWORD) (PJob []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumJobsRequest{
		HPrinter: hPrinter,
		FirstJob: firstJob,
		NoJobs:   noJobs,
		Level:    level,
		PJob:     pJob,
		CbBuf:    cbBuf,
	}
	var resp rpcEnumJobsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumJobs: %w", err)
		return
	}
	PJob = resp.PJob
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumJobs failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
