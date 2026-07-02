package mswcce

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CERTTRANSBLOB models the CERTTRANSBLOB structure ([MS-WCCE] 2.2.2.2), a
// generic byte buffer carried by the certificate-services RPC interfaces (for
// example MS-WCCE's ICertRequestD and MS-ICPR's ICertPassage).
//
// Cb is the count of bytes in Pb. Pb is a [size_is(cb)] pointer to that
// conformant byte array; under pointer_default(unique) it is a unique pointer,
// so the wire form is a referent id followed (when non-null) by the conformant
// array body.
type CERTTRANSBLOB struct {
	Cb ndr.DWORD
	Pb []uint8 `ndr:"unique,size_is=Cb"`
}
