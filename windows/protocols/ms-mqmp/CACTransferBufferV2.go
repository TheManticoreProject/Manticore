package msmqmp

import (
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// CACTransferBufferV2 extends CACTransferBufferV1 with the transaction fields of a Send or
// Receive operation ([MS-MQMP] 2.2.6): the embedded V1 buffer plus first/last-in-transaction
// flags and the transaction identifier.
//
// NOT YET WIRE-READY. It embeds CACTransferBufferV1, so it inherits that type's
// double-indirection limitation (see the note there) and is likewise deferred and not
// round-trip tested. PpXactID is the IDL "OBJECTID**", modelled here with a single
// indirection.
type CACTransferBufferV2 struct {
	Old           CACTransferBufferV1
	PbFirstInXact *uint8           `ndr:"unique"`
	PbLastInXact  *uint8           `ndr:"unique"`
	PpXactID      *msmqmq.OBJECTID `ndr:"unique"`
}
