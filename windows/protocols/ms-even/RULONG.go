package mseven

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// RULONG is a range-restricted unsigned long ([MS-EVEN] 2.2.4:
// [range(0, MAX_BATCH_BUFF)] unsigned long). It is the type of the
// NumberOfBytesToRead argument to ElfrReadELW/ElfrReadELA. On the wire it is a
// plain 4-octet NDR unsigned long; the range is a server-side validation bound.
type RULONG ndr.DWORD
