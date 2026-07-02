package mscmrp

// CLUSDSK_DISKID_ENUM is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-CMRP]).
type CLUSDSK_DISKID_ENUM uint16

const (
	DiskIdSignature CLUSDSK_DISKID_ENUM = 0x00000001
	DiskIdGuid      CLUSDSK_DISKID_ENUM = 0x00000002
	DiskIdUnKnown   CLUSDSK_DISKID_ENUM = 0x00001388
)
