package structures

import "github.com/TheManticoreProject/Manticore/windows/guid"

// RpcIfID models RPC_IF_ID ([C706] Appendix O / [MS-RPCE] 2.2.1.2.4): an interface
// identifier (UUID + 16-bit major/minor version). It is the referent of ept_lookup's
// optional [in, ptr] Ifid parameter, used to filter the endpoint map by interface.
//
// UUID is the EptUUID octet form (the same fixed-octet UUID used by ept_map's object
// pointer) so the codec emits it verbatim as part of the referent body.
type RpcIfID struct {
	UUID      EptUUID
	VersMajor uint16
	VersMinor uint16
}

// NewRpcIfID builds an RpcIfID from a GUID and an interface major/minor version.
func NewRpcIfID(iface guid.GUID, major, minor uint16) RpcIfID {
	return RpcIfID{UUID: NewEptUUID(iface), VersMajor: major, VersMinor: minor}
}
