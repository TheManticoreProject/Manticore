package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// OBJECTID uniquely distinguishes objects of the same type within Message Queuing
// ([MS-MQMQ] 2.2.8): a GUID group identifier (Lineage) and a DWORD object identifier
// (Uniquifier). It is the m_oPrivateID arm of QUEUE_FORMAT (a private queue) and is
// reused by [MS-MQMP] as a message identifier.
type OBJECTID struct {
	Lineage    msdtyp.GUID
	Uniquifier ndr.DWORD
}
