package msmqds

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// MQPROPERTYRESTRICTION is a single property restriction in a query ([MS-MQDS] 2.2.11): a
// relational operator (rel), the property to test (prop), and the value to compare against
// (prval). The value is an [MS-MQMQ] PROPVARIANT.
type MQPROPERTYRESTRICTION struct {
	Rel   ndr.DWORD
	Prop  ndr.DWORD
	Prval msmqmq.PROPVARIANT
}
