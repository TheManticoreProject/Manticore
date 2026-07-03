package msmqmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// MGMT_OBJECT identifies the management object a request targets ([MS-MQMR] 2.2.2): a
// type discriminant followed by a [switch_is(type)] union carrying the per-kind
// identifier. The MGMT_QUEUE arm is a unique pointer to a QUEUE_FORMAT; the MGMT_MACHINE
// and MGMT_SESSION arms are each a reserved DWORD (its value is unused by the server, but
// the 4 bytes are still transmitted as the selected arm) ([MS-MQMR] 2.2.2).
//
// Wire modelling. The IDL union is non-encapsulated (its discriminant is the enclosing
// struct's type field). Per [C706] 14.3.8 a non-encapsulated union transmits its
// discriminant twice — once as the enclosing type field and once as part of the union
// representation — and the codec's declarative unions carry their own discriminant. So
// the union is modelled as the Value field whose own Type switch mirrors the outer Type;
// keep the two equal with SetType. This follows the QUEUE_FORMAT/PROPVARIANT convention
// in the ms-mqmq package.
type MGMT_OBJECT struct {
	Type  MgmtObjectType `ndr:"enum"`
	Value MgmtObjectUnion
}

// SetType sets both the outer discriminant and the union's mirrored discriminant so they
// stay consistent, then returns the receiver for chaining.
func (m *MGMT_OBJECT) SetType(t MgmtObjectType) *MGMT_OBJECT {
	m.Type = t
	m.Value.Type = t
	return m
}

// MgmtObjectUnion is the [switch_is(type)] union of MGMT_OBJECT ([MS-MQMR] 2.2.2). The
// discriminant is an NDR enum ([C706] 14.3.6), so it carries the "enum" tag: the codec
// emits it at the syntax width (2 octets under NDR20, 4 under NDR64), not as a bare
// uint16. The MGMT_QUEUE arm is a unique pointer to a QUEUE_FORMAT ([MS-MQMQ] 2.2.7),
// matching pointer_default(unique).
type MgmtObjectUnion struct {
	Type MgmtObjectType `ndr:"switch,enum"`

	Reserved1    ndr.DWORD            `ndr:"case=1"`        // MGMT_MACHINE
	PQueueFormat *msmqmq.QUEUE_FORMAT `ndr:"case=2,unique"` // MGMT_QUEUE
	Reserved2    ndr.DWORD            `ndr:"case=3"`        // MGMT_SESSION
}
