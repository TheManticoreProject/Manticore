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

// resolveOxid2Request carries the [in] parameters of ResolveOxid2. See resolveOxidRequest
// for the pointer/array modeling notes; the parameters are identical to ResolveOxid.
type resolveOxid2Request struct {
	POxid               msdcom.OXID
	CRequestedProtseqs  uint16
	ArRequestedProtseqs []uint16 `ndr:"ref,size_is=CRequestedProtseqs"`
}

func (*resolveOxid2Request) Opnum() uint16 { return IObjectExporter.OpnumResolveOxid2 }

// resolveOxid2Response carries the [out] parameters and return value of ResolveOxid2. It
// adds the object exporter's COMVERSION to the ResolveOxid result; ppdsaOxidBindings is
// modeled as *DUALSTRINGARRAY "unique" for the same reason as in ResolveOxid.
type resolveOxid2Response struct {
	PpdsaOxidBindings *msdcom.DUALSTRINGARRAY `ndr:"unique"`
	PipidRemUnknown   msdcom.IPID
	PAuthnHint        ndr.DWORD
	PComVersion       msdcom.COMVERSION
	Status            ndr.DWORD `ndr:"retval"`
}

// ResolveOxid2 calls ResolveOxid2 (opnum 4) ([MS-DCOM] 3.1.2.5.1.2): like ResolveOxid, but
// it also returns the COMVERSION of the object exporter.
func ResolveOxid2(rpc ndr.Invoker, pOxid msdcom.OXID, cRequestedProtseqs uint16, arRequestedProtseqs []uint16) (PpdsaOxidBindings *msdcom.DUALSTRINGARRAY, PipidRemUnknown msdcom.IPID, PAuthnHint ndr.DWORD, PComVersion msdcom.COMVERSION, err error) {
	req := &resolveOxid2Request{
		POxid:               pOxid,
		CRequestedProtseqs:  cRequestedProtseqs,
		ArRequestedProtseqs: arRequestedProtseqs,
	}
	var resp resolveOxid2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ResolveOxid2: %w", err)
		return
	}
	PpdsaOxidBindings = resp.PpdsaOxidBindings
	PipidRemUnknown = resp.PipidRemUnknown
	PAuthnHint = resp.PAuthnHint
	PComVersion = resp.PComVersion
	if uint32(resp.Status) != IObjectExporter.StatusSuccess {
		err = fmt.Errorf("ResolveOxid2 failed: %s", IObjectExporter.StatusString(uint32(resp.Status)))
	}
	return
}
