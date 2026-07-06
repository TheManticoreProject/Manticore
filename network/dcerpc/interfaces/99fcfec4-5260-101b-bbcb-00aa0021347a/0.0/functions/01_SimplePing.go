package functions

// IDL source: [MS-DCOM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dcom/49aef5a4-f0ad-4478-abb5-cb9446dc13c6
// A fetched copy is kept at ms-dcom.idl in the interface directory.

import (
	"fmt"

	IObjectExporter "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/99fcfec4-5260-101b-bbcb-00aa0021347a/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdcom "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dcom"
)

// simplePingRequest carries the [in] parameters of SimplePing. pSetId is a [ref] pointer
// to a SETID and is transmitted inline.
type simplePingRequest struct {
	PSetId msdcom.SETID
}

func (*simplePingRequest) Opnum() uint16 { return IObjectExporter.OpnumSimplePing }

// simplePingResponse carries the [out] parameters and return value of SimplePing.
type simplePingResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// SimplePing calls SimplePing (opnum 1) ([MS-DCOM] 3.1.2.5.1.5): it pings the object
// resolver ping set identified by pSetId to keep its objects alive. The set must already
// exist (created by ComplexPing); SimplePing cannot modify set membership.
func SimplePing(rpc ndr.Invoker, pSetId msdcom.SETID) (err error) {
	req := &simplePingRequest{
		PSetId: pSetId,
	}
	var resp simplePingResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("SimplePing: %w", err)
		return
	}
	if uint32(resp.Status) != IObjectExporter.StatusSuccess {
		err = fmt.Errorf("SimplePing failed: %s", IObjectExporter.StatusString(uint32(resp.Status)))
	}
	return
}
