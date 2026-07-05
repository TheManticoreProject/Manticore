package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrShareDelCommitRequest is the [in] parameter set of NetrShareDelCommit: the [in,out]
// SHARE_DEL_HANDLE context handle obtained from NetrShareDelStart.
type netrShareDelCommitRequest struct {
	ContextHandle mssrvs.SHARE_DEL_HANDLE
}

func (*netrShareDelCommitRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareDelCommit
}

// netrShareDelCommitResponse is the reply: the [in,out] SHARE_DEL_HANDLE context handle and
// the NET_API_STATUS return value.
type netrShareDelCommitResponse struct {
	ContextHandle mssrvs.SHARE_DEL_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// NetrShareDelCommit calls NetrShareDelCommit (opnum 38), completing the two-phase share
// delete started by NetrShareDelStart ([MS-SRVS] 3.1.4.17).
func NetrShareDelCommit(rpc ndr.Invoker, contextHandle mssrvs.SHARE_DEL_HANDLE) (mssrvs.SHARE_DEL_HANDLE, error) {
	req := &netrShareDelCommitRequest{
		ContextHandle: contextHandle,
	}
	var resp netrShareDelCommitResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssrvs.SHARE_DEL_HANDLE{}, fmt.Errorf("NetrShareDelCommit: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.ContextHandle, fmt.Errorf("NetrShareDelCommit failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.ContextHandle, nil
}
