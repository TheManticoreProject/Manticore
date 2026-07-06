package msrrasm

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MIB_OPAQUE_INFO ([MS-RRASM] 2.2.1.2.207) is a variable-length opaque MIB data
// container. In the IDL the trailing field is the C union
// {ULONGLONG ullAlign; BYTE rgbyData[1]}, which forces the payload to an
// 8-byte-aligned offset. UllAlign models that fixed 8-byte alignment slot; the
// actual rgbyData payload (whose layout depends on DwId) follows it in the buffer.
type MIB_OPAQUE_INFO struct {
	DwId     ndr.DWORD
	UllAlign uint64
}
