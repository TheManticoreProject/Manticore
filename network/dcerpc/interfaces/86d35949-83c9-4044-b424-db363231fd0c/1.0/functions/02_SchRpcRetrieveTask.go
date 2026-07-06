package functions

// IDL source: [MS-TSCH] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsch/6fc1f51a-26ec-43fa-a8bd-1c364657f110
// A fetched copy is kept at ms-tsch.idl in the interface directory.

import (
	"fmt"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// schRpcRetrieveTaskRequest carries the [in] parameters of SchRpcRetrieveTask.
type schRpcRetrieveTaskRequest struct {
	Path                  ndr.WSTR
	LpcwszLanguagesBuffer ndr.WSTR
	PulNumLanguages       ndr.DWORD
}

func (*schRpcRetrieveTaskRequest) Opnum() uint16 {
	return schrpc.OpnumSchRpcRetrieveTask
}

// schRpcRetrieveTaskResponse carries the [out] parameters and return value of SchRpcRetrieveTask.
type schRpcRetrieveTaskResponse struct {
	PXml   *ndr.WSTR `ndr:"unique"`
	Status ndr.DWORD `ndr:"retval"`
}

// SchRpcRetrieveTask calls SchRpcRetrieveTask (opnum 2) ([MS-TSCH] section 3.2.5.4.3).
func SchRpcRetrieveTask(rpc ndr.Invoker, path ndr.WSTR, lpcwszLanguagesBuffer ndr.WSTR, pulNumLanguages ndr.DWORD) (PXml *ndr.WSTR, err error) {
	req := &schRpcRetrieveTaskRequest{
		Path:                  path,
		LpcwszLanguagesBuffer: lpcwszLanguagesBuffer,
		PulNumLanguages:       pulNumLanguages,
	}
	var resp schRpcRetrieveTaskResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SchRpcRetrieveTask: %w", err)
		return
	}
	PXml = resp.PXml
	if !schrpc.IsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("SchRpcRetrieveTask failed: %s", schrpc.StatusString(uint32(resp.Status)))
	}
	return
}
