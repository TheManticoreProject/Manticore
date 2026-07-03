package msmqmp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// OBJECT_FORMAT identifies a directory object operated on by [MS-MQMP] ([MS-MQMP] 2.2.3):
// an ObjType discriminant and a [switch_is(ObjType)] union. ObjType 1 selects a queue,
// carried by a unique pointer to a QUEUE_FORMAT ([MS-MQMQ] 2.2.7).
//
// Wire modelling. The IDL union is non-encapsulated (its discriminant is the enclosing
// struct's ObjType field); the codec's declarative unions carry their own discriminant, so
// the union is modelled as the Value field whose own ObjType switch mirrors the outer one.
// Keep the two equal with SetObjType.
type OBJECT_FORMAT struct {
	ObjType ndr.DWORD
	Value   ObjectFormatUnion
}

// SetObjType sets both the outer discriminant and the union's mirrored discriminant so
// they stay consistent, then returns the receiver for chaining.
func (o *OBJECT_FORMAT) SetObjType(objType ndr.DWORD) *OBJECT_FORMAT {
	o.ObjType = objType
	o.Value.ObjType = objType
	return o
}

// ObjectFormatUnion is the [switch_is(ObjType)] union of OBJECT_FORMAT ([MS-MQMP] 2.2.3).
// ObjType 1 selects a unique pointer to a QUEUE_FORMAT; ObjType 2 selects no arm (an empty
// union body).
type ObjectFormatUnion struct {
	ObjType      ndr.DWORD            `ndr:"switch"`
	PQueueFormat *msmqmq.QUEUE_FORMAT `ndr:"case=1,unique"`
}
