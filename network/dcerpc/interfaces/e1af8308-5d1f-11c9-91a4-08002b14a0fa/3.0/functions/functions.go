// Package functions holds the endpoint mapper (ept) method stubs. Each stub depends
// only on the small ndr.Invoker surface, so it is independent of any concrete client or
// wire-protocol version.
package functions

// IDL source: [C706] — this interface is translated from and verified
// against the protocol's authoritative IDL. Authoritative IDL reference:
//   https://pubs.opengroup.org/onlinepubs/9629399/apdxo.htm
// No standalone MS-* Full IDL page exists; the reference above is authoritative.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msrpce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpce"
)

// DefaultMaxTowers is the number of towers Map asks ept_map to return. A handful covers
// the endpoints a single interface is typically bound to.
const DefaultMaxTowers = 4

// DefaultMaxEnts is the batch size Lookup requests per ept_lookup call. ept_lookup caps
// max_ents at 500 ([MS-RPCE] 2.2.1.2.4 range(0,500)); enumeration pages until the entry
// handle is exhausted regardless of this value.
const DefaultMaxEnts = 500

// Map resolves the ncacn_ip_tcp endpoints bound to the given interface UUID and version
// by building a TCP map tower and calling ept_map, then extracting the endpoints from
// the returned towers. It is the common path for discovering the dynamic TCP port a
// service listens on; for finer control (a non-nil object, a custom tower, or the raw
// towers) call EptMap directly.
func Map(rpc ndr.Invoker, iface guid.GUID, ifMajor, ifMinor uint16) ([]msrpce.Endpoint, error) {
	towers, err := EptMap(rpc, nil, msrpce.BuildMapTowerTCP(iface, ifMajor, ifMinor), DefaultMaxTowers)
	if err != nil {
		return nil, err
	}
	var eps []msrpce.Endpoint
	for _, t := range towers {
		if ep, ok := t.Endpoint(); ok {
			eps = append(eps, ep)
		}
	}
	return eps, nil
}
