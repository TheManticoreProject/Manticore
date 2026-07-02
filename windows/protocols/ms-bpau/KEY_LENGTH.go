package msbpau

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// KEY_LENGTH is the length in bytes of a marshalled CERTIFICATE_BLOB key buffer
// ([MS-BPAU] 2.2.1). The IDL declares it as a [range(0, 65536)] DWORD, so it is a
// 32-bit unsigned value bounded to [0, 65536]; a value of zero means no key is present.
type KEY_LENGTH ndr.DWORD

// KeyLengthMax is the inclusive upper bound of the [range(0, 65536)] constraint the IDL
// places on KEY_LENGTH ([MS-BPAU] 2.2.1).
const KeyLengthMax KEY_LENGTH = 65536
