package functions

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// beginTearDownRequest carries the [in] parameters of BeginTearDown ([MS-CMPO] 3.4.4.6).
type beginTearDownRequest struct {
	ContextHandle mscmpo.PCONTEXT_HANDLE
	TearDownType  mscmpo.TEARDOWN_TYPE `ndr:"enum"`
}

func (*beginTearDownRequest) Opnum() uint16 { return IXnRemote.OpnumBeginTearDown }

// beginTearDownResponse carries the HRESULT return value of BeginTearDown.
type beginTearDownResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BeginTearDown calls BeginTearDown (opnum 5) ([MS-CMPO] 3.4.4.6): it signals the partner
// to begin tearing down the session identified by contextHandle.
func BeginTearDown(rpc ndr.Invoker, contextHandle mscmpo.PCONTEXT_HANDLE, tearDownType mscmpo.TEARDOWN_TYPE) error {
	req := &beginTearDownRequest{
		ContextHandle: contextHandle,
		TearDownType:  tearDownType,
	}
	var resp beginTearDownResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("BeginTearDown: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return fmt.Errorf("BeginTearDown failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return nil
}
