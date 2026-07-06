package functions

// IDL source: [MS-EFSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
// A fetched copy is kept at ms-efsr.idl in the interface directory.

import (
	"fmt"

	efsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/c681d488-d850-11d0-8c52-00c04fd90f7e/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msefsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-efsr"
)

// efsRpcReadFileRawRequest carries the [in] parameters of EfsRpcReadFileRaw: the
// export context handle opened by EfsRpcOpenFileRaw.
type efsRpcReadFileRawRequest struct {
	HContext msefsr.PEXIMPORT_CONTEXT_HANDLE
}

func (*efsRpcReadFileRawRequest) Opnum() uint16 { return efsrpc.OpnumEfsRpcReadFileRaw }

// efsRpcReadFileRawResponse carries the [out] EFS_EXIM_PIPE and the return value.
// EfsOutPipe is an NDR pipe ([C706] 14.7) — the server streams the raw encrypted file
// to the client as chunks ahead of the return value.
type efsRpcReadFileRawResponse struct {
	EfsOutPipe msefsr.EFS_EXIM_PIPE `ndr:"pipe"`
	Status     ndr.DWORD            `ndr:"retval"`
}

// EfsRpcReadFileRaw calls EfsRpcReadFileRaw (opnum 1, [MS-EFSR] 3.1.4.2.2). It exports
// the opened file as a raw encrypted stream, returning the full pipe contents.
func EfsRpcReadFileRaw(rpc ndr.Invoker, hContext msefsr.PEXIMPORT_CONTEXT_HANDLE) (msefsr.EFS_EXIM_PIPE, error) {
	req := &efsRpcReadFileRawRequest{HContext: hContext}
	var resp efsRpcReadFileRawResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("EfsRpcReadFileRaw: %w", err)
	}
	if uint32(resp.Status) != efsrpc.StatusSuccess {
		return resp.EfsOutPipe, fmt.Errorf("EfsRpcReadFileRaw failed: %s", efsrpc.StatusString(uint32(resp.Status)))
	}
	return resp.EfsOutPipe, nil
}
