package msfrs2

import "github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"

// FRS_MEMBER_ID is a GUID that identifies a replication member ([MS-FRS2] 2.2.1.4). It aliases dtyp.GUID (the
// 16-octet NDR wire form), not windows/guid.GUID (whose trailing uint64 marshals to 24
// octets), so it serializes correctly on the wire.
type FRS_MEMBER_ID = dtyp.GUID
