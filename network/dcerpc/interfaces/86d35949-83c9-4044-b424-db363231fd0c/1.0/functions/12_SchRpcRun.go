package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// schRpcRunRequest carries the [in] parameters of SchRpcRun.
type schRpcRunRequest struct {
	Path      ndr.WSTR
	CArgs     ndr.DWORD
	PArgs     []*ndr.WSTR `ndr:"unique,size_is=CArgs"`
	Flags     ndr.DWORD
	SessionId ndr.DWORD
	User      *ndr.WSTR `ndr:"unique"`
}

func (*schRpcRunRequest) Opnum() uint16 { return schrpc.OpnumSchRpcRun }

// schRpcRunResponse carries the [out] parameters and return value of SchRpcRun.
type schRpcRunResponse struct {
	PGuid  guid.GUID
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcRun calls SchRpcRun (opnum 12) ([MS-TSCH] section 3.2.5.4.13).
func SchRpcRun(rpc ndr.Invoker, path ndr.WSTR, cArgs ndr.DWORD, pArgs []*ndr.WSTR, flags ndr.DWORD, sessionId ndr.DWORD, user *ndr.WSTR) (PGuid guid.GUID, err error) {
	req := &schRpcRunRequest{
		Path:      path,
		CArgs:     cArgs,
		PArgs:     pArgs,
		Flags:     flags,
		SessionId: sessionId,
		User:      user,
	}
	var resp schRpcRunResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcRun: %w", err)
		return
	}
	PGuid = resp.PGuid
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcRun failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
