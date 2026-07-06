package msfrs2

import msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"

// FRS_REPLICA_SET_ID is a GUID that identifies a replication group ([MS-FRS2] 2.2.1.1). It aliases msdtyp.GUID (the
// 16-octet NDR wire form), not windows/guid.GUID (whose trailing uint64 marshals to 24
// octets), so it serializes correctly on the wire.
type FRS_REPLICA_SET_ID = msdtyp.GUID
