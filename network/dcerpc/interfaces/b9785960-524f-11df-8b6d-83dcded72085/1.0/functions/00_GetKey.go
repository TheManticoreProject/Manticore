package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	ISDKey "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b9785960-524f-11df-8b6d-83dcded72085/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// getKeyRequest carries the [in] parameters of GetKey.
//
// PRootKeyID is [unique] GUID* (nullable) — the NDR-marshallable dtyp.GUID
// (16 octets on the wire), NOT windows/guid.GUID whose uint64 tail would
// marshal to 24 octets. A nil pointer requests the latest / time-derived key.
type getKeyRequest struct {
	CbTargetSD ndr.DWORD
	PbTargetSD []byte     `ndr:"ref,size_is=CbTargetSD"`
	PRootKeyID *dtyp.GUID `ndr:"unique"`
	L0KeyID    int32
	L1KeyID    int32
	L2KeyID    int32
}

func (*getKeyRequest) Opnum() uint16 { return ISDKey.OpnumGetKey }

// getKeyResponse carries the [out] parameters and return value of GetKey.
//
// ppbOut is [out][size_is(, *pcbOut)] byte** — a [unique] pointer to a
// conformant array of bytes (the key BLOB of section 2.2.4), NOT an array of
// pointers. Its conformant count is *pcbOut. The HRESULT return value is
// encoded after this deferred referent, hence the retval tag.
type getKeyResponse struct {
	PcbOut ndr.DWORD
	PpbOut []byte    `ndr:"unique,size_is=PcbOut"`
	Status ndr.DWORD `ndr:"retval"`
}

// GetKey calls GetKey (opnum 0) ([MS-GKDI] section 3.1.4.1). It requests a
// group key (seed or public) for the security descriptor pbTargetSD. Setting
// pRootKeyID to nil and l0/l1/l2KeyID all to -1 requests the latest group key;
// all three >= 0 requests a specific seed key. On success it returns the key
// BLOB (section 2.2.4) and its length. The wire return value is an HRESULT
// (zero on success, nonzero on error).
func GetKey(rpc ndr.Invoker, cbTargetSD ndr.DWORD, pbTargetSD []byte, pRootKeyID *guid.GUID, l0KeyID int32, l1KeyID int32, l2KeyID int32) (PcbOut ndr.DWORD, PpbOut []byte, err error) {
	var rootKeyID *dtyp.GUID
	if pRootKeyID != nil {
		g := dtyp.NewGUID(*pRootKeyID)
		rootKeyID = &g
	}
	req := &getKeyRequest{
		CbTargetSD: cbTargetSD,
		PbTargetSD: pbTargetSD,
		PRootKeyID: rootKeyID,
		L0KeyID:    l0KeyID,
		L1KeyID:    l1KeyID,
		L2KeyID:    l2KeyID,
	}
	var resp getKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("GetKey: %w", err)
		return
	}
	PcbOut = resp.PcbOut
	PpbOut = resp.PpbOut
	if uint32(resp.Status) != ISDKey.StatusSuccess {
		err = fmt.Errorf("GetKey failed: %s", ISDKey.StatusString(uint32(resp.Status)))
	}
	return
}
