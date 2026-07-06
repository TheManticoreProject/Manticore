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

// schRpcGetInstanceInfoRequest carries the [in] parameters of SchRpcGetInstanceInfo.
type schRpcGetInstanceInfoRequest struct {
	Guid guid.GUID
}

func (*schRpcGetInstanceInfoRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcGetInstanceInfo
}

// schRpcGetInstanceInfoResponse carries the [out] parameters and return value of SchRpcGetInstanceInfo.
type schRpcGetInstanceInfoResponse struct {
	PPath            *ndr.WSTR `ndr:"unique"`
	PState           ndr.DWORD
	PCurrentAction   *ndr.WSTR `ndr:"unique"`
	PInfo            *ndr.WSTR `ndr:"unique"`
	PcGroupInstances ndr.DWORD
	PGroupInstances  []guid.GUID `ndr:"unique,size_is=PcGroupInstances"`
	PEnginePID       ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// SchRpcGetInstanceInfo calls SchRpcGetInstanceInfo (opnum 9) ([MS-TSCH] section 3.2.5.4.10).
func SchRpcGetInstanceInfo(rpc ndr.Invoker, guid guid.GUID) (PPath *ndr.WSTR, PState ndr.DWORD, PCurrentAction *ndr.WSTR, PInfo *ndr.WSTR, PcGroupInstances ndr.DWORD, PGroupInstances []guid.GUID, PEnginePID ndr.DWORD, err error) {
	req := &schRpcGetInstanceInfoRequest{
		Guid: guid,
	}
	var resp schRpcGetInstanceInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcGetInstanceInfo: %w", err)
		return
	}
	PPath = resp.PPath
	PState = resp.PState
	PCurrentAction = resp.PCurrentAction
	PInfo = resp.PInfo
	PcGroupInstances = resp.PcGroupInstances
	PGroupInstances = resp.PGroupInstances
	PEnginePID = resp.PEnginePID
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcGetInstanceInfo failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
