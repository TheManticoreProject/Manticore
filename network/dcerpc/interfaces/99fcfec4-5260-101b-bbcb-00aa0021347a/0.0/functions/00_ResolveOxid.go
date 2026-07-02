package functions

import (
	"fmt"

	IObjectExporter "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/99fcfec4-5260-101b-bbcb-00aa0021347a/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdcom "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dcom"
)

// resolveOxidRequest carries the [in] parameters of ResolveOxid. The explicit binding
// handle (handle_t hRpc) is not part of the NDR stub and is omitted. pOxid is a [ref]
// pointer to an OXID, so it is transmitted inline; arRequestedProtseqs is a [ref] pointer
// to a conformant array of protocol sequence identifiers (its maximum_count travels with
// the array, no referent id).
type resolveOxidRequest struct {
	POxid               msdcom.OXID
	CRequestedProtseqs  uint16
	ArRequestedProtseqs []uint16 `ndr:"ref,size_is=CRequestedProtseqs"`
}

func (*resolveOxidRequest) Opnum() uint16 { return IObjectExporter.OpnumResolveOxid }

// resolveOxidResponse carries the [out] parameters and return value of ResolveOxid.
// ppdsaOxidBindings is [out, ref] DUALSTRINGARRAY **: the outer [ref] is the by-reference
// out slot (inline), and the inner pointer takes the interface's unique default, so the
// field is a *DUALSTRINGARRAY tagged "unique" (referent id, then the deferred body).
type resolveOxidResponse struct {
	PpdsaOxidBindings *msdcom.DUALSTRINGARRAY `ndr:"unique"`
	PipidRemUnknown   msdcom.IPID
	PAuthnHint        ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// ResolveOxid calls ResolveOxid (opnum 0) ([MS-DCOM] 3.1.2.5.1.1): it resolves an OXID to
// the string and security bindings of the object exporter and the IPID of that exporter's
// IRemUnknown interface.
func ResolveOxid(rpc ndr.Invoker, pOxid msdcom.OXID, cRequestedProtseqs uint16, arRequestedProtseqs []uint16) (PpdsaOxidBindings *msdcom.DUALSTRINGARRAY, PipidRemUnknown msdcom.IPID, PAuthnHint ndr.DWORD, err error) {
	req := &resolveOxidRequest{
		POxid:               pOxid,
		CRequestedProtseqs:  cRequestedProtseqs,
		ArRequestedProtseqs: arRequestedProtseqs,
	}
	var resp resolveOxidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ResolveOxid: %w", err)
		return
	}
	PpdsaOxidBindings = resp.PpdsaOxidBindings
	PipidRemUnknown = resp.PipidRemUnknown
	PAuthnHint = resp.PAuthnHint
	if uint32(resp.Status) != IObjectExporter.StatusSuccess {
		err = fmt.Errorf("ResolveOxid failed: %s", IObjectExporter.StatusString(uint32(resp.Status)))
	}
	return
}
