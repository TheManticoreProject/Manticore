package functions

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// tearDownContextRequest carries the [in]/[in,out] parameters of TearDownContext
// ([MS-CMPO] 3.4.4.5). contextHandle is [in,out]: on success the server returns the nulled
// handle, so it also appears in the response.
type tearDownContextRequest struct {
	ContextHandle mscmpo.PPCONTEXT_HANDLE
	SRank         mscmpo.SESSION_RANK  `ndr:"enum"`
	TearDownType  mscmpo.TEARDOWN_TYPE `ndr:"enum"`
}

func (*tearDownContextRequest) Opnum() uint16 { return IXnRemote.OpnumTearDownContext }

// tearDownContextResponse carries the [in,out] (nulled) handle and the HRESULT return
// value of TearDownContext.
type tearDownContextResponse struct {
	ContextHandle mscmpo.PPCONTEXT_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// TearDownContext calls TearDownContext (opnum 4) ([MS-CMPO] 3.4.4.5): it tears down the
// session identified by contextHandle. It returns the (typically nulled) handle the
// server sends back.
func TearDownContext(rpc ndr.Invoker, contextHandle mscmpo.PPCONTEXT_HANDLE, sRank mscmpo.SESSION_RANK, tearDownType mscmpo.TEARDOWN_TYPE) (mscmpo.PPCONTEXT_HANDLE, error) {
	req := &tearDownContextRequest{
		ContextHandle: contextHandle,
		SRank:         sRank,
		TearDownType:  tearDownType,
	}
	var resp tearDownContextResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mscmpo.PPCONTEXT_HANDLE{}, fmt.Errorf("TearDownContext: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return resp.ContextHandle, fmt.Errorf("TearDownContext failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return resp.ContextHandle, nil
}
