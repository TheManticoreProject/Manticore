package mscapr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsat "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsat"
)

// LSAPR_WRAPPED_CAPID_SET is the wrapper returned by LsarGetAvailableCAPIDs
// ([MS-CAPR] 2.2.1.1). Entries is the number of central access policy objects, and
// SidInfo is a [unique] pointer to a conformant array of that many
// LSAPR_SID_INFORMATION structures (each holds one CAPID, that is, the SID of a
// central access policy object).
type LSAPR_WRAPPED_CAPID_SET struct {
	Entries ndr.DWORD
	SidInfo []mslsat.LSAPR_SID_INFORMATION `ndr:"unique,size_is=Entries"`
}
