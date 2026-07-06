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

// complexPingRequest carries the [in]/[in,out] parameters of ComplexPing. pSetId is a
// [ref] pointer to a SETID (transmitted inline); addToSet and delFromSet are [unique]
// pointers to conformant arrays of OIDs, so each carries a referent id followed by the
// deferred array body (maximum_count + elements) when non-null.
type complexPingRequest struct {
	PSetId      msdcom.SETID
	SequenceNum uint16
	CAddToSet   uint16
	CDelFromSet uint16
	AddToSet    []msdcom.OID `ndr:"unique,size_is=CAddToSet"`
	DelFromSet  []msdcom.OID `ndr:"unique,size_is=CDelFromSet"`
}

func (*complexPingRequest) Opnum() uint16 { return IObjectExporter.OpnumComplexPing }

// complexPingResponse carries the [in,out] and [out] parameters and return value of
// ComplexPing. pSetId is the (possibly server-assigned) ping set id; it is [in,out] so it
// appears in both the request and the response.
type complexPingResponse struct {
	PSetId             msdcom.SETID
	PPingBackoffFactor uint16
	Status             ndr.DWORD `ndr:"retval"`
}

// ComplexPing calls ComplexPing (opnum 2) ([MS-DCOM] 3.1.2.5.1.6): it maintains an object
// resolver ping set, adding and removing OIDs from the set identified by pSetId (zero on
// the first call, whereupon the server assigns and returns a new set id). Keeping objects
// in a pinged set prevents the server from reclaiming them.
func ComplexPing(rpc ndr.Invoker, pSetId msdcom.SETID, sequenceNum uint16, cAddToSet uint16, cDelFromSet uint16, addToSet []msdcom.OID, delFromSet []msdcom.OID) (PSetId msdcom.SETID, PPingBackoffFactor uint16, err error) {
	req := &complexPingRequest{
		PSetId:      pSetId,
		SequenceNum: sequenceNum,
		CAddToSet:   cAddToSet,
		CDelFromSet: cDelFromSet,
		AddToSet:    addToSet,
		DelFromSet:  delFromSet,
	}
	var resp complexPingResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ComplexPing: %w", err)
		return
	}
	PSetId = resp.PSetId
	PPingBackoffFactor = resp.PPingBackoffFactor
	if uint32(resp.Status) != IObjectExporter.StatusSuccess {
		err = fmt.Errorf("ComplexPing failed: %s", IObjectExporter.StatusString(uint32(resp.Status)))
	}
	return
}
