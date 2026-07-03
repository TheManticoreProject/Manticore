package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// QUEUE_FORMAT_TYPE enumerates the queue format-name kinds ([MS-MQMQ] 2.2.7.1). Its
// values are the m_qft discriminant of QUEUE_FORMAT; m_qft is transmitted as an
// unsigned char, so the constants are modelled as byte values.
type QUEUE_FORMAT_TYPE = uint8

const (
	QUEUE_FORMAT_TYPE_UNKNOWN   QUEUE_FORMAT_TYPE = 0
	QUEUE_FORMAT_TYPE_PUBLIC    QUEUE_FORMAT_TYPE = 1
	QUEUE_FORMAT_TYPE_PRIVATE   QUEUE_FORMAT_TYPE = 2
	QUEUE_FORMAT_TYPE_DIRECT    QUEUE_FORMAT_TYPE = 3
	QUEUE_FORMAT_TYPE_MACHINE   QUEUE_FORMAT_TYPE = 4
	QUEUE_FORMAT_TYPE_CONNECTOR QUEUE_FORMAT_TYPE = 5
	QUEUE_FORMAT_TYPE_DL        QUEUE_FORMAT_TYPE = 6
	QUEUE_FORMAT_TYPE_MULTICAST QUEUE_FORMAT_TYPE = 7
	QUEUE_FORMAT_TYPE_SUBQUEUE  QUEUE_FORMAT_TYPE = 8
)

// QUEUE_FORMAT describes the type of a queue and an identifier for it ([MS-MQMQ] 2.2.7):
// an m_qft discriminant, an m_SuffixAndFlags byte (Suffix Type in the low nibble, Flags in
// the high nibble), a reserved padding word, and a [switch_is(m_qft)] union carrying the
// per-kind identifier.
//
// Wire modelling. The IDL union is non-encapsulated (its discriminant is the enclosing
// struct's m_qft field). The codec's declarative unions carry their own discriminant, so
// the union is modelled as the Value field whose own MQft switch mirrors the outer MQft;
// keep the two equal with SetQft. This follows the same convention as PROPVARIANT in this
// package.
type QUEUE_FORMAT struct {
	MQft            QUEUE_FORMAT_TYPE
	MSuffixAndFlags uint8
	MReserved       uint16
	Value           QueueFormatUnion
}

// SetQft sets both the outer discriminant and the union's mirrored discriminant so they
// stay consistent, then returns the receiver for chaining.
func (q *QUEUE_FORMAT) SetQft(qft QUEUE_FORMAT_TYPE) *QUEUE_FORMAT {
	q.MQft = qft
	q.Value.MQft = qft
	return q
}

// QueueFormatUnion is the [switch_is(m_qft)] union of QUEUE_FORMAT ([MS-MQMQ] 2.2.7). The
// UNKNOWN discriminant (0) selects no arm (an empty union body). The direct/subqueue arms
// are unique pointers to wide strings, matching pointer_default(unique).
type QueueFormatUnion struct {
	MQft QUEUE_FORMAT_TYPE `ndr:"switch"`

	MGPublicID         dtyp.GUID    `ndr:"case=1"`        // QUEUE_FORMAT_TYPE_PUBLIC
	MOPrivateID        OBJECTID     `ndr:"case=2"`        // QUEUE_FORMAT_TYPE_PRIVATE
	MPDirectID         *ndr.WSTR    `ndr:"case=3,unique"` // QUEUE_FORMAT_TYPE_DIRECT
	MGMachineID        dtyp.GUID    `ndr:"case=4"`        // QUEUE_FORMAT_TYPE_MACHINE
	MGConnectorID      dtyp.GUID    `ndr:"case=5"`        // QUEUE_FORMAT_TYPE_CONNECTOR
	MDlID              DL_ID        `ndr:"case=6"`        // QUEUE_FORMAT_TYPE_DL
	MMulticastID       MULTICAST_ID `ndr:"case=7"`        // QUEUE_FORMAT_TYPE_MULTICAST
	MPDirectSubqueueID *ndr.WSTR    `ndr:"case=8,unique"` // QUEUE_FORMAT_TYPE_SUBQUEUE
}
