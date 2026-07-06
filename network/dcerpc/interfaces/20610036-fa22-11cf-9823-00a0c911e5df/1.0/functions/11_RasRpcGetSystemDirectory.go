package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rasRpcGetSystemDirectoryRequest carries the [in] parameters of RasRpcGetSystemDirectory.
type rasRpcGetSystemDirectoryRequest struct {
	LpBuffer ndr.WSTR
	USize    uint32
}

func (*rasRpcGetSystemDirectoryRequest) Opnum() uint16 { return rasrpc.OpnumRasRpcGetSystemDirectory }

// rasRpcGetSystemDirectoryResponse carries the [out] parameters and return value of RasRpcGetSystemDirectory.
type rasRpcGetSystemDirectoryResponse struct {
	LpBuffer ndr.WSTR
	Status   ndr.DWORD `ndr:"retval"`
}

// RasRpcGetSystemDirectory calls RasRpcGetSystemDirectory (opnum 11) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcGetSystemDirectory(rpc ndr.Invoker, lpBuffer ndr.WSTR, uSize uint32) (LpBuffer ndr.WSTR, err error) {
	req := &rasRpcGetSystemDirectoryRequest{
		LpBuffer: lpBuffer,
		USize:    uSize,
	}
	var resp rasRpcGetSystemDirectoryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcGetSystemDirectory: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcGetSystemDirectory failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
