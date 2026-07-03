package functions

import (
	"fmt"

	IRPCRemoteObject "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ae33069b-a2a8-46ee-a235-ddfd339be281/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCRemoteObject_CreateRequest carries the [in] parameters of IRPCRemoteObject_Create.
// The only [in] parameter is the explicit handle_t binding handle, which is not
// marshalled into the request stub, so the request body is empty.
type iRPCRemoteObject_CreateRequest struct {
}

func (*iRPCRemoteObject_CreateRequest) Opnum() uint16 {
	return IRPCRemoteObject.OpnumIRPCRemoteObject_Create
}

// iRPCRemoteObject_CreateResponse carries the [out] parameters and return value of IRPCRemoteObject_Create.
type iRPCRemoteObject_CreateResponse struct {
	PpRemoteObj mspan.PRPCREMOTEOBJECT
	Status      ndr.DWORD `ndr:"retval"`
}

// IRPCRemoteObject_Create calls IRPCRemoteObject_Create (opnum 0) ([MS-PAN] 3.1.2.4.1). It
// creates a new remote object on the server and returns its context handle for use as the
// registration/channel argument of the IRPCAsyncNotify methods.
func IRPCRemoteObject_Create(rpc ndr.Invoker) (PpRemoteObj mspan.PRPCREMOTEOBJECT, err error) {
	req := &iRPCRemoteObject_CreateRequest{}
	var resp iRPCRemoteObject_CreateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCRemoteObject_Create: %w", err)
		return
	}
	PpRemoteObj = resp.PpRemoteObj
	if uint32(resp.Status) != IRPCRemoteObject.StatusSuccess {
		err = fmt.Errorf("IRPCRemoteObject_Create failed: %s", IRPCRemoteObject.StatusString(uint32(resp.Status)))
	}
	return
}
