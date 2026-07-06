package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncEnumJobsRequest carries the [in] parameters of RpcAsyncEnumJobs.
type rpcAsyncEnumJobsRequest struct {
	HPrinter mspar.PRINTER_HANDLE
	FirstJob ndr.DWORD
	NoJobs   ndr.DWORD
	Level    ndr.DWORD
	PJob     []uint8 `ndr:"ref,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAsyncEnumJobsRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncEnumJobs }

// rpcAsyncEnumJobsResponse carries the [out] parameters and return value of RpcAsyncEnumJobs.
type rpcAsyncEnumJobsResponse struct {
	PJob       []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded  ndr.DWORD
	PcReturned ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumJobs calls RpcAsyncEnumJobs (opnum 4) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumJobs(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, firstJob ndr.DWORD, noJobs ndr.DWORD, level ndr.DWORD, pJob []uint8, cbBuf ndr.DWORD) (PJob []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumJobsRequest{
		HPrinter: hPrinter,
		FirstJob: firstJob,
		NoJobs:   noJobs,
		Level:    level,
		PJob:     pJob,
		CbBuf:    cbBuf,
	}
	var resp rpcAsyncEnumJobsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumJobs: %w", err)
		return
	}
	PJob = resp.PJob
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumJobs failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
