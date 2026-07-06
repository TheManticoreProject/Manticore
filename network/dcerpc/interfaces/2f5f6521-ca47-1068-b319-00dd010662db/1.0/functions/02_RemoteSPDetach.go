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

// remoteSPDetachRequest carries the [in,out] context handle of RemoteSPDetach, transmitted
// inline as a [ref] pointer to the 20-octet context handle.
type remoteSPDetachRequest struct {
	PphContext mstrp.PCONTEXT_HANDLE_TYPE2
}

func (*remoteSPDetachRequest) Opnum() uint16 { return remotesp.OpnumRemoteSPDetach }

// remoteSPDetachResponse carries the [in,out] context handle back. RemoteSPDetach is a
// void method, so there is no return value on the wire; the callee returns the handle
// nulled out on success.
type remoteSPDetachResponse struct {
	PphContext mstrp.PCONTEXT_HANDLE_TYPE2
}

// RemoteSPDetach calls RemoteSPDetach (opnum 2) ([MS-TRP] 3.1.4.3). The telephony server
// invokes it on the client to tear down the reverse binding established by RemoteSPAttach
// and free the associated state. The callee returns the handle nulled out (IsZero);
// transport-level failures surface through err.
func RemoteSPDetach(rpc ndr.Invoker, pphContext mstrp.PCONTEXT_HANDLE_TYPE2) (PphContext mstrp.PCONTEXT_HANDLE_TYPE2, err error) {
	req := &remoteSPDetachRequest{
		PphContext: pphContext,
	}
	var resp remoteSPDetachResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteSPDetach: %w", err)
		return
	}
	PphContext = resp.PphContext
	return
}
