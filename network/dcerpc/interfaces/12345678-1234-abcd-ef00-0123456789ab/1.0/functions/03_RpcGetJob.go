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

// rpcGetJobRequest carries the [in] parameters of RpcGetJob.
type rpcGetJobRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	JobId    ndr.DWORD
	Level    ndr.DWORD
	PJob     []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcGetJobRequest) Opnum() uint16 { return winspool.OpnumRpcGetJob }

// rpcGetJobResponse carries the [out] parameters and return value of RpcGetJob.
type rpcGetJobResponse struct {
	PJob      []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcGetJob calls RpcGetJob (opnum 3) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetJob(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, jobId ndr.DWORD, level ndr.DWORD, pJob []uint8, cbBuf ndr.DWORD) (PJob []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetJobRequest{
		HPrinter: hPrinter,
		JobId:    jobId,
		Level:    level,
		PJob:     pJob,
		CbBuf:    cbBuf,
	}
	var resp rpcGetJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetJob: %w", err)
		return
	}
	PJob = resp.PJob
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetJob failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
