package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseNode failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
