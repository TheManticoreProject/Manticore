package functions

// IDL source: [MS-TRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-trp/e86aca98-76e9-4515-9de1-2cadb9084a2b
// A fetched copy is kept at ms-trp.idl in the interface directory.

import (
	"fmt"

	tapsrv "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/2f5f6520-ca46-1067-b319-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-trp"
)

// clientDetachRequest carries the [in,out] context handle of ClientDetach. The handle is
// transmitted inline as a [ref] pointer to the 20-octet context handle.
type clientDetachRequest struct {
	PphContext mstrp.PCONTEXT_HANDLE_TYPE
}

func (*clientDetachRequest) Opnum() uint16 { return tapsrv.OpnumClientDetach }

// clientDetachResponse carries the [in,out] context handle back. ClientDetach is a void
// method, so there is no return value on the wire; the server returns the handle nulled
// out on success.
type clientDetachResponse struct {
	PphContext mstrp.PCONTEXT_HANDLE_TYPE
}

// ClientDetach calls ClientDetach (opnum 2) ([MS-TRP] 3.2.4.3). It closes the tapsrv
// session identified by the context handle and frees the server-side state. The server
// returns the handle nulled out (IsZero); transport-level failures surface through err.
func ClientDetach(rpc ndr.Invoker, pphContext mstrp.PCONTEXT_HANDLE_TYPE) (PphContext mstrp.PCONTEXT_HANDLE_TYPE, err error) {
	req := &clientDetachRequest{
		PphContext: pphContext,
	}
	var resp clientDetachResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ClientDetach: %w", err)
		return
	}
	PphContext = resp.PphContext
	return
}
