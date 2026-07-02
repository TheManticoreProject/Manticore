package functions

import (
	"fmt"

	IObjectExporter "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/99fcfec4-5260-101b-bbcb-00aa0021347a/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdcom "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dcom"
)

// serverAlive2Request carries the [in] parameters of ServerAlive2.
type serverAlive2Request struct {
}

func (*serverAlive2Request) Opnum() uint16 { return IObjectExporter.OpnumServerAlive2 }

// serverAlive2Response carries the [out] parameters and return value of ServerAlive2.
// ppdsaOrBindings is [out, ref] DUALSTRINGARRAY ** — the string/security bindings the
// object resolver supports — modeled as *DUALSTRINGARRAY "unique" (see ResolveOxid).
type serverAlive2Response struct {
	PComVersion     msdcom.COMVERSION
	PpdsaOrBindings *msdcom.DUALSTRINGARRAY `ndr:"unique"`
	PReserved       ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// ServerAlive2 calls ServerAlive2 (opnum 5) ([MS-DCOM] 3.1.2.5.1.4): it tests whether the
// object resolver is alive and returns its COMVERSION and the string/security bindings it
// supports. It is the method a client uses to discover a working RPC protocol sequence
// ([MS-DCOM] 2.1).
func ServerAlive2(rpc ndr.Invoker) (PComVersion msdcom.COMVERSION, PpdsaOrBindings *msdcom.DUALSTRINGARRAY, PReserved ndr.DWORD, err error) {
	req := &serverAlive2Request{}
	var resp serverAlive2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ServerAlive2: %w", err)
		return
	}
	PComVersion = resp.PComVersion
	PpdsaOrBindings = resp.PpdsaOrBindings
	PReserved = resp.PReserved
	if uint32(resp.Status) != IObjectExporter.StatusSuccess {
		err = fmt.Errorf("ServerAlive2 failed: %s", IObjectExporter.StatusString(uint32(resp.Status)))
	}
	return
}
