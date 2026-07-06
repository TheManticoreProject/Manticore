package functions

// IDL source: [C706] — this interface is translated from and verified
// against the protocol's authoritative IDL. Authoritative IDL reference:
//   https://pubs.opengroup.org/onlinepubs/9629399/apdxo.htm
// No standalone MS-* Full IDL page exists; the reference above is authoritative.

import (
	"fmt"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msrpce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpce"
)

// eptMapRequest carries the [in] parameters of ept_map. The explicit binding handle
// (handle_t h) is not part of the NDR stub and is omitted. The optional object is a
// full pointer to a uuid_t; map_tower is a full pointer to a twr_t; entry_handle is the
// 20-octet context handle (null on the first call); max_towers caps the result.
type eptMapRequest struct {
	Object      *msrpce.EptUUID `ndr:"ptr"`
	MapTower    *msrpce.Twr     `ndr:"ptr"`
	EntryHandle msrpce.ContextHandle
	MaxTowers   ndr.DWORD
}

func (*eptMapRequest) Opnum() uint16 { return epm.OpnumEptMap }

// eptMapResponse carries the [out] parameters of ept_map: the (advanced) context handle,
// the number of towers returned, the towers themselves, and the status.
//
// In the IDL ([C706] Appendix O; [MS-RPCE] 2.2.1.2.5) ITowers is a bare, non-pointer,
// top-level conformant-varying array of full pointers to twr_t:
//
//	[out, length_is(*num_towers), size_is(max_towers)] twr_p_t ITowers[]
//
// The array itself is not pointer-prefixed — only its elements (twr_p_t) are pointers. So
// its conformance (maximum_count) is transmitted inline, immediately after num_towers,
// with no referent id and without being hoisted to the front of the parameter structure;
// each element then carries its own referent id with the twr_t body deferred. The "inline"
// tag selects that framing, "varying" supplies the offset/actual_count words, and
// "elem=ptr" makes the elements full pointers.
type eptMapResponse struct {
	EntryHandle msrpce.ContextHandle
	NumTowers   ndr.DWORD
	ITowers     []*msrpce.Twr `ndr:"varying,inline,elem=ptr"`
	Status      ndr.DWORD
}

// EptMap calls ept_map (opnum 3) ([C706] Appendix O): it asks the endpoint mapper to
// resolve the map tower to the towers of the matching bound endpoints. object is the
// optional object UUID (nil for the usual interface lookup); maxTowers caps how many
// towers the server returns. The decoded result towers are returned; use Tower.Endpoint
// to extract a transport endpoint, or call Map for the common interface-to-endpoint
// path.
func EptMap(rpc ndr.Invoker, object *guid.GUID, mapTower msrpce.Tower, maxTowers uint32) ([]msrpce.Tower, error) {
	twr := msrpce.NewTwr(mapTower)
	req := &eptMapRequest{
		MapTower:  &twr,
		MaxTowers: ndr.DWORD(maxTowers),
	}
	if object != nil {
		u := msrpce.NewEptUUID(*object)
		req.Object = &u
	}

	var resp eptMapResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("ept_map: %w", err)
	}
	if uint32(resp.Status) != epm.EptStatusSuccess {
		return nil, fmt.Errorf("ept_map failed: %s", epm.StatusString(uint32(resp.Status)))
	}

	towers := make([]msrpce.Tower, 0, len(resp.ITowers))
	for _, tw := range resp.ITowers {
		if tw == nil {
			continue
		}
		t, err := tw.DecodeTower()
		if err != nil {
			return nil, fmt.Errorf("ept_map: decode tower: %w", err)
		}
		towers = append(towers, t)
	}
	return towers, nil
}
