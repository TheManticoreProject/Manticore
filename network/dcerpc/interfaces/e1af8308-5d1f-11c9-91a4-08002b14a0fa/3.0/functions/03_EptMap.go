package functions

import (
	"fmt"

	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// eptMapRequest carries the [in] parameters of ept_map. The explicit binding handle
// (handle_t h) is not part of the NDR stub and is omitted. The optional object is a
// full pointer to a uuid_t; map_tower is a full pointer to a twr_t; entry_handle is the
// 20-octet context handle (null on the first call); max_towers caps the result.
type eptMapRequest struct {
	Object      *structures.EptUUID `ndr:"ptr"`
	MapTower    *structures.Twr     `ndr:"ptr"`
	EntryHandle structures.ContextHandle
	MaxTowers   ndr.DWORD
}

func (*eptMapRequest) Opnum() uint16 { return epm.OpnumEptMap }

// eptMapResponse carries the [out] parameters of ept_map: the (advanced) context
// handle, the number of towers returned, the towers themselves (a full pointer to a
// conformant-varying array of full pointers to twr_t), and the status.
type eptMapResponse struct {
	EntryHandle structures.ContextHandle
	NumTowers   ndr.DWORD
	ITowers     []*structures.Twr `ndr:"ptr,varying,elem=ptr"`
	Status      ndr.DWORD
}

// EptMap calls ept_map (opnum 3) ([C706] Appendix O): it asks the endpoint mapper to
// resolve the map tower to the towers of the matching bound endpoints. object is the
// optional object UUID (nil for the usual interface lookup); maxTowers caps how many
// towers the server returns. The decoded result towers are returned; use Tower.Endpoint
// to extract a transport endpoint, or call Map for the common interface-to-endpoint
// path.
func EptMap(rpc ndr.Invoker, object *guid.GUID, mapTower structures.Tower, maxTowers uint32) ([]structures.Tower, error) {
	twr := structures.NewTwr(mapTower)
	req := &eptMapRequest{
		MapTower:  &twr,
		MaxTowers: ndr.DWORD(maxTowers),
	}
	if object != nil {
		u := structures.NewEptUUID(*object)
		req.Object = &u
	}

	var resp eptMapResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("ept_map: %w", err)
	}
	if uint32(resp.Status) != epm.StatusSuccess {
		return nil, fmt.Errorf("ept_map failed: %s", epm.StatusString(uint32(resp.Status)))
	}

	towers := make([]structures.Tower, 0, len(resp.ITowers))
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
