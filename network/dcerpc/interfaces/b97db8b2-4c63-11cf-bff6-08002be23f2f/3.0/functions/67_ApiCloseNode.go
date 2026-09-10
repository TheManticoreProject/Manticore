package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCloseNodeRequest carries the [in] parameters of ApiCloseNode.
type apiCloseNodeRequest struct {
	Node mscmrp.HNODE_RPC
}

func (*apiCloseNodeRequest) Opnum() uint16 { return clusapi.OpnumApiCloseNode }

// apiCloseNodeResponse carries the [out] parameters and return value of ApiCloseNode.
type apiCloseNodeResponse struct {
	Node   mscmrp.HNODE_RPC
	Status ndr.DWORD `ndr:"retval"`
}

// ApiCloseNode calls ApiCloseNode (opnum 67) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseNode(rpc ndr.Invoker, node mscmrp.HNODE_RPC) (Node mscmrp.HNODE_RPC, err error) {
	req := &apiCloseNodeRequest{
		Node: node,
	}
	var resp apiCloseNodeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseNode: %w", err)
		return
	}
	Node = resp.Node
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("ApiCloseNode failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
