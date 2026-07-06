package msfrs2

import msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"

// FRS_CONNECTION_ID is a GUID that identifies a connection ([MS-FRS2] 2.2.1.5). It aliases msdtyp.GUID (the
// 16-octet NDR wire form), not windows/guid.GUID (whose trailing uint64 marshals to 24
// octets), so it serializes correctly on the wire.
type FRS_CONNECTION_ID = msdtyp.GUID
