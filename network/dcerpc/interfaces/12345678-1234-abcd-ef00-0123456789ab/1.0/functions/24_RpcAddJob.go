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

// rpcAddJobRequest carries the [in] parameters of RpcAddJob.
type rpcAddJobRequest struct {
	HPrinter msrprn.PRINTER_HANDLE
	Level    ndr.DWORD
	PAddJob  []uint8 `ndr:"unique,size_is=CbBuf"`
	CbBuf    ndr.DWORD
}

func (*rpcAddJobRequest) Opnum() uint16 { return winspool.OpnumRpcAddJob }

// rpcAddJobResponse carries the [out] parameters and return value of RpcAddJob.
type rpcAddJobResponse struct {
	PAddJob   []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAddJob calls RpcAddJob (opnum 24) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddJob(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, level ndr.DWORD, pAddJob []uint8, cbBuf ndr.DWORD) (PAddJob []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAddJobRequest{
		HPrinter: hPrinter,
		Level:    level,
		PAddJob:  pAddJob,
		CbBuf:    cbBuf,
	}
	var resp rpcAddJobResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddJob: %w", err)
		return
	}
	PAddJob = resp.PAddJob
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddJob failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
