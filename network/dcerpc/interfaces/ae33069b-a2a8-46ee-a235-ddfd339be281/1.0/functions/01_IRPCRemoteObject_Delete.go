package functions

import (
	"fmt"

	IRPCRemoteObject "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ae33069b-a2a8-46ee-a235-ddfd339be281/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCRemoteObject_DeleteRequest carries the [in] parameters of IRPCRemoteObject_Delete.
type iRPCRemoteObject_DeleteRequest struct {
	PpRemoteObj mspan.PRPCREMOTEOBJECT
}

func (*iRPCRemoteObject_DeleteRequest) Opnum() uint16 {
	return IRPCRemoteObject.OpnumIRPCRemoteObject_Delete
}

// iRPCRemoteObject_DeleteResponse carries the [in,out] parameter of IRPCRemoteObject_Delete.
// The method returns void ([MS-PAN] 3.1.2.4.2), so there is no HRESULT on the wire; the
// server sets ppRemoteObj to NULL on return.
type iRPCRemoteObject_DeleteResponse struct {
	PpRemoteObj mspan.PRPCREMOTEOBJECT
}

// IRPCRemoteObject_Delete calls IRPCRemoteObject_Delete (opnum 1) ([MS-PAN] 3.1.2.4.2). It
// destroys the remote object referenced by ppRemoteObj; the returned handle is the
// server-nulled [in,out] context handle. The method has no return value.
func IRPCRemoteObject_Delete(rpc ndr.Invoker, ppRemoteObj mspan.PRPCREMOTEOBJECT) (PpRemoteObj mspan.PRPCREMOTEOBJECT, err error) {
	req := &iRPCRemoteObject_DeleteRequest{
		PpRemoteObj: ppRemoteObj,
	}
	var resp iRPCRemoteObject_DeleteResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCRemoteObject_Delete: %w", err)
		return
	}
	PpRemoteObj = resp.PpRemoteObj
	return
}
