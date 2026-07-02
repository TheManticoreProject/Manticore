package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_CR_CIPHER_VALUE carries an encrypted (or cleartext) secret value ([MS-LSAD]
// 2.2.6.1). Buffer is a [unique] pointer to a conformant-varying byte array whose
// maximum_count is MaximumLength and actual_count is Length.
type LSAPR_CR_CIPHER_VALUE struct {
	Length        ndr.DWORD
	MaximumLength ndr.DWORD
	Buffer        []byte `ndr:"unique,varying,size_is=MaximumLength,length_is=Length"`
}
