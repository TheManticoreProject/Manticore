package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// DL_ID identifies a distribution list ([MS-MQMQ] 2.2.9): a GUID and a unique pointer to
// the domain name string. It is the m_DlID arm of QUEUE_FORMAT.
type DL_ID struct {
	MDlGuid    msdtyp.GUID
	MPwzDomain *ndr.WSTR `ndr:"unique"`
}
