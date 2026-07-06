package functions

// IDL source: [MS-TRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-trp/e86aca98-76e9-4515-9de1-2cadb9084a2b
// A fetched copy is kept at ms-trp.idl in the interface directory.

import (
	"fmt"

	remotesp "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/2f5f6521-ca47-1068-b319-00dd010662db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// remoteSPAttachRequest carries the (empty) [in] parameter list of RemoteSPAttach.
type remoteSPAttachRequest struct {
}

func (*remoteSPAttachRequest) Opnum() uint16 { return remotesp.OpnumRemoteSPAttach }

// remoteSPAttachResponse carries the [out] context handle and return value of
// RemoteSPAttach.
type remoteSPAttachResponse struct {
	PphContext mstrp.PCONTEXT_HANDLE_TYPE2
	Return     ndr.DWORD `ndr:"retval"`
}

// RemoteSPAttach calls RemoteSPAttach (opnum 0) ([MS-TRP] 3.1.4.1). In the protocol the
// telephony server is the RPC client for the remotesp interface: it calls RemoteSPAttach
// on the client (which hosts the remotesp server) to establish a reverse binding for
// event delivery. On success the callee returns the remotesp context handle. The method
// returns 0 on success or a nonzero error code.
func RemoteSPAttach(rpc ndr.Invoker) (PphContext mstrp.PCONTEXT_HANDLE_TYPE2, err error) {
	req := &remoteSPAttachRequest{}
	var resp remoteSPAttachResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteSPAttach: %w", err)
		return
	}
	PphContext = resp.PphContext
	if uint32(resp.Return) != remotesp.StatusSuccess {
		err = fmt.Errorf("RemoteSPAttach failed: %s", remotesp.StatusString(uint32(resp.Return)))
	}
	return
}
