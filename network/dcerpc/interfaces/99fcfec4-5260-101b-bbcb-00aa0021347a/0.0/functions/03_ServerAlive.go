package functions

// IDL source: [MS-DCOM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dcom/49aef5a4-f0ad-4478-abb5-cb9446dc13c6
// A fetched copy is kept at ms-dcom.idl in the interface directory.

import (
	"fmt"

	IObjectExporter "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/99fcfec4-5260-101b-bbcb-00aa0021347a/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// serverAliveRequest carries the [in] parameters of ServerAlive.
type serverAliveRequest struct {
}

func (*serverAliveRequest) Opnum() uint16 { return IObjectExporter.OpnumServerAlive }

// serverAliveResponse carries the [out] parameters and return value of ServerAlive.
type serverAliveResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// ServerAlive calls ServerAlive (opnum 3) ([MS-DCOM] 3.1.2.5.1.3): it tests whether the
// object resolver is alive and reachable. It takes no marshaled parameters and returns
// only a status; ServerAlive2 is preferred as it also reports version and bindings.
func ServerAlive(rpc ndr.Invoker) (err error) {
	req := &serverAliveRequest{}
	var resp serverAliveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ServerAlive: %w", err)
		return
	}
	if uint32(resp.Status) != IObjectExporter.StatusSuccess {
		err = fmt.Errorf("ServerAlive failed: %s", IObjectExporter.StatusString(uint32(resp.Status)))
	}
	return
}
