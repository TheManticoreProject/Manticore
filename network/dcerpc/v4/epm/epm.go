package epm

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Endpoint mapper (ept) interface identity ([C706] Appendix O).
var (
	// Interface is the ept interface UUID, e1af8308-5d1f-11c9-91a4-08002b14a0fa.
	Interface = guid.GUID{A: 0xe1af8308, B: 0x5d1f, C: 0x11c9, D: 0x91a4, E: 0x08002b14a0fa}
)

const (
	// InterfaceMajorVersion is the ept interface major version (3.0).
	InterfaceMajorVersion = 3
	// InterfaceMinorVersion is the ept interface minor version.
	InterfaceMinorVersion = 0
	// OpnumEptMap is the operation number of ept_map (the 4th ept operation).
	OpnumEptMap = 3
	// DefaultMaxTowers is the default number of towers ept_map is asked to return.
	DefaultMaxTowers = 4
)

// NDR referent IDs for the non-null full pointers in the ept_map request. Any
// distinct non-zero values are valid; these follow the conventional NDR scheme.
const (
	refObject   = 0x00000001
	refMapTower = 0x00020000
)

// contextHandleSize is the wire size of an ept_lookup_handle_t context handle:
// 4-octet attributes plus a 16-octet UUID.
const contextHandleSize = 20

// Client resolves interface endpoints through the endpoint mapper using a
// connectionless RPC client.
type Client struct {
	rpc       *client.Client
	maxTowers uint32
}

// New returns an endpoint-mapper client over the given connectionless RPC client.
// The rpc client should be bound to the endpoint mapper's UDP endpoint (port 135).
func New(rpc *client.Client) *Client {
	return &Client{rpc: rpc, maxTowers: DefaultMaxTowers}
}

// SetMaxTowers overrides how many towers ept_map is asked to return.
func (c *Client) SetMaxTowers(n uint32) { c.maxTowers = n }

// Map resolves the transport endpoints bound to the given interface UUID and version
// by calling ept_map with an ncadg_ip_udp map tower, and returns the endpoints
// extracted from the towers the endpoint mapper returns.
func (c *Client) Map(iface guid.GUID, ifMajor, ifMinor uint16) ([]Endpoint, error) {
	tower := BuildMapTower(iface, ifMajor, ifMinor)
	stub := marshalEptMapRequest(nil, tower, c.maxTowers)

	resp, err := c.rpc.Call(client.CallRequest{
		Interface:        Interface,
		InterfaceVersion: InterfaceMajorVersion,
		OpNum:            OpnumEptMap,
		Stub:             stub,
		Idempotent:       true, // ept_map is declared [idempotent]
	})
	if err != nil {
		return nil, err
	}

	towers, status, err := parseEptMapResponse(resp)
	if err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, fmt.Errorf("epm: ept_map returned status 0x%08x", status)
	}

	var eps []Endpoint
	for _, t := range towers {
		if ep, ok := t.Endpoint(); ok {
			eps = append(eps, ep)
		}
	}
	return eps, nil
}

// marshalEptMapRequest builds the NDR stub for the ept_map [in] parameters:
//
//	uuid_p_t  object       (full pointer; null when object is nil)
//	twr_p_t   map_tower    (full pointer to a twr_t)
//	ept_lookup_handle_t entry_handle (a 20-octet context handle, null on input)
//	unsigned32 max_towers
//
// The twr_t pointee is { unsigned32 tower_length; [size_is] byte tower[] }; as a
// conformant struct its array maximum_count is hoisted to the front of the struct.
func marshalEptMapRequest(object *guid.GUID, mapTower Tower, maxTowers uint32) []byte {
	w := &ndrWriter{}

	// object: full pointer, null unless provided.
	if object == nil {
		w.u32(0)
	} else {
		w.u32(refObject)
		w.bytes(object.ToBytes())
	}

	// map_tower: non-null full pointer to a twr_t.
	w.u32(refMapTower)
	towerBytes := mapTower.Marshal()
	w.u32(uint32(len(towerBytes))) // hoisted conformant maximum_count
	w.u32(uint32(len(towerBytes))) // tower_length
	w.bytes(towerBytes)
	w.align(4)

	// entry_handle: null context handle (all zero).
	w.bytes(make([]byte, contextHandleSize))

	// max_towers.
	w.u32(maxTowers)

	return w.buf
}

// maxResponseTowers bounds how many tower referents parseEptMapResponse will accept,
// guarding against a malformed count driving a large allocation.
const maxResponseTowers = 4096

// parseEptMapResponse decodes the ept_map [out] parameters from a response stub:
//
//	ept_lookup_handle_t entry_handle (20 octets, skipped)
//	unsigned32 num_towers
//	twr_p_t towers[]  (conformant+varying array of full pointers)
//	error_status_t status
//
// It returns the successfully decoded towers and the status code.
func parseEptMapResponse(data []byte) ([]Tower, uint32, error) {
	r := &ndrReader{data: data}

	r.skip(contextHandleSize) // entry_handle
	_ = r.u32()               // num_towers (the array's actual_count below is authoritative)

	// towers[]: conformant+varying array header.
	_ = r.u32() // maximum_count
	_ = r.u32() // offset
	actualCount := r.u32()
	if r.err() != nil {
		return nil, 0, r.err()
	}
	if actualCount > maxResponseTowers {
		return nil, 0, fmt.Errorf("epm: ept_map returned implausible tower count %d", actualCount)
	}

	// Referent IDs for each array element, in order.
	refs := make([]uint32, actualCount)
	for i := range refs {
		refs[i] = r.u32()
	}
	if r.err() != nil {
		return nil, 0, r.err()
	}

	// Deferred twr_t pointees, one per non-null referent.
	towers := make([]Tower, 0, actualCount)
	for _, ref := range refs {
		if ref == 0 {
			continue
		}
		_ = r.u32() // hoisted conformant maximum_count
		towerLen := r.u32()
		raw := r.take(int(towerLen))
		r.align(4)
		if r.err() != nil {
			return nil, 0, r.err()
		}
		t, err := UnmarshalTower(raw)
		if err != nil {
			return nil, 0, err
		}
		towers = append(towers, t)
	}

	status := r.u32()
	if r.err() != nil {
		return nil, 0, r.err()
	}
	return towers, status, nil
}
