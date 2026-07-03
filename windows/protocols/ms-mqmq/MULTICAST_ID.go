package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// MULTICAST_ID identifies a multicast queue ([MS-MQMQ] 2.2.10): an IP address and the
// port the queue is attached to. It is the m_MulticastID arm of QUEUE_FORMAT.
type MULTICAST_ID struct {
	MAddress ndr.DWORD // ULONG m_address
	MPort    ndr.DWORD // ULONG m_port
}
