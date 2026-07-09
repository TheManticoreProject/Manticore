// Package messages provides Kerberos protocol message types and constants
// as defined in RFC 4120 and related specifications.
package messages

import "github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"

// The protocol constants live canonically in the iana leaf package. They are
// re-exported here as aliases so message code and existing callers can keep
// referring to messages.<Const>; iana remains the single source of truth.
const (
	KerberosV5 = iana.KerberosV5

	MsgTypeASReq   = iana.MsgTypeASReq
	MsgTypeASRep   = iana.MsgTypeASRep
	MsgTypeTGSReq  = iana.MsgTypeTGSReq
	MsgTypeTGSRep  = iana.MsgTypeTGSRep
	MsgTypeAPReq   = iana.MsgTypeAPReq
	MsgTypeAPRep   = iana.MsgTypeAPRep
	MsgTypeKRBCred = iana.MsgTypeKRBCred
	MsgTypeError   = iana.MsgTypeError

	NameTypePrincipal  = iana.NameTypePrincipal
	NameTypeSRVInst    = iana.NameTypeSRVInst
	NameTypeSRVHST     = iana.NameTypeSRVHST
	NameTypeEnterprise = iana.NameTypeEnterprise

	ETypeRC4HMAC             = iana.ETypeRC4HMAC
	ETypeAES128CTSHMACSHA196 = iana.ETypeAES128CTSHMACSHA196
	ETypeAES256CTSHMACSHA196 = iana.ETypeAES256CTSHMACSHA196

	PATGSReq       = iana.PATGSReq
	PAEncTimestamp = iana.PAEncTimestamp
	PAETypeInfo2   = iana.PAETypeInfo2
	PAPACRequest   = iana.PAPACRequest

	ErrNone              = iana.ErrNone
	ErrCPrincipalUnknown = iana.ErrCPrincipalUnknown
	ErrPreauthRequired   = iana.ErrPreauthRequired

	APOptionUseSessionKey = iana.APOptionUseSessionKey
	APOptionMutualAuth    = iana.APOptionMutualAuth

	TicketFlagForwardable = iana.TicketFlagForwardable
	TicketFlagForwarded   = iana.TicketFlagForwarded
	TicketFlagProxiable   = iana.TicketFlagProxiable
	TicketFlagProxy       = iana.TicketFlagProxy
	TicketFlagPreAuthent  = iana.TicketFlagPreAuthent
	TicketFlagInitial     = iana.TicketFlagInitial
	TicketFlagRenewable   = iana.TicketFlagRenewable
)
