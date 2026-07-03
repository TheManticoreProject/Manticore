package msnrpc

// SYNC_STATE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-NRPC]).
type SYNC_STATE uint16

const (
	NormalState          SYNC_STATE = 0
	DomainState          SYNC_STATE = 1
	GroupState           SYNC_STATE = 2
	UasBuiltInGroupState SYNC_STATE = 3
	UserState            SYNC_STATE = 4
	GroupMemberState     SYNC_STATE = 5
	AliasState           SYNC_STATE = 6
	AliasMemberState     SYNC_STATE = 7
	SamDoneState         SYNC_STATE = 8
)
