package msmqmr

// MgmtObjectType is an NDR enum ([C706] 14.3.6, [MS-MQMR] 2.2.2). Its Go base type is
// uint16, but on the wire it is an enum: fields of this type must carry the ndr "enum"
// tag so the codec emits the syntax-correct width (2 octets under NDR20, 4 under NDR64).
type MgmtObjectType uint16

const (
	MGMT_MACHINE MgmtObjectType = 1
	MGMT_QUEUE   MgmtObjectType = 2
	MGMT_SESSION MgmtObjectType = 3
)
