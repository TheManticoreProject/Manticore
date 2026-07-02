package msfrs2

import "github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"

// FRS_DATABASE_ID is a GUID that identifies a database ([MS-FRS2] 2.2.1.3). It aliases dtyp.GUID (the
// 16-octet NDR wire form), not windows/guid.GUID (whose trailing uint64 marshals to 24
// octets), so it serializes correctly on the wire.
type FRS_DATABASE_ID = dtyp.GUID
