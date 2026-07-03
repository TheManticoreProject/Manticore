package functions

import (
	"fmt"

	IRPCAsyncNotify "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b6edbfa-4a24-4fc6-8a23-942b1eca65d1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCAsyncNotify_UnregisterClientRequest carries the [in] parameters of IRPCAsyncNotify_UnregisterClient.
type iRPCAsyncNotify_UnregisterClientRequest struct {
	PRegistrationObj mspan.PRPCREMOTEOBJECT
}

func (*iRPCAsyncNotify_UnregisterClientRequest) Opnum() uint16 {
	return IRPCAsyncNotify.OpnumIRPCAsyncNotify_UnregisterClient
}

// iRPCAsyncNotify_UnregisterClientResponse carries the [out] parameters and return value of IRPCAsyncNotify_UnregisterClient.
type iRPCAsyncNotify_UnregisterClientResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IRPCAsyncNotify_UnregisterClient calls IRPCAsyncNotify_UnregisterClient (opnum 1)
// ([MS-PAN] 3.1.4.2). It cancels a prior successful registration for the given remote
// object.
func IRPCAsyncNotify_UnregisterClient(rpc ndr.Invoker, pRegistrationObj mspan.PRPCREMOTEOBJECT) (err error) {
	req := &iRPCAsyncNotify_UnregisterClientRequest{
		PRegistrationObj: pRegistrationObj,
	}
	var resp iRPCAsyncNotify_UnregisterClientResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCAsyncNotify_UnregisterClient: %w", err)
		return
	}
	if uint32(resp.Status) != IRPCAsyncNotify.StatusSuccess {
		err = fmt.Errorf("IRPCAsyncNotify_UnregisterClient failed: %s", IRPCAsyncNotify.StatusString(uint32(resp.Status)))
	}
	return
}
