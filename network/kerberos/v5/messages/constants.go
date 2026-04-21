// Package messages provides Kerberos protocol message types and constants
// as defined in RFC 4120 and related specifications.
package messages

// Kerberos message type constants (RFC 4120 Section 7.5.7).
const (
	// MsgTypeASReq is the Authentication Service Request message type.
	MsgTypeASReq = 10
	// MsgTypeASRep is the Authentication Service Reply message type.
	MsgTypeASRep = 11
	// MsgTypeTGSReq is the Ticket Granting Service Request message type.
	MsgTypeTGSReq = 12
	// MsgTypeTGSRep is the Ticket Granting Service Reply message type.
	MsgTypeTGSRep = 13
	// MsgTypeAPReq is the Application Request message type.
	MsgTypeAPReq = 14
	// MsgTypeAPRep is the Application Reply message type.
	MsgTypeAPRep = 15
	// MsgTypeError is the KRB-ERROR message type.
	MsgTypeError = 30
)

// KerberosV5 is the Kerberos protocol version number.
const KerberosV5 = 5

// Principal name type constants (RFC 4120 Section 6.2).
const (
	// NameTypePrincipal is the general principal name type (NT-PRINCIPAL).
	NameTypePrincipal = 1
	// NameTypeSRVInst is the service instance name type (NT-SRV-INST).
	NameTypeSRVInst = 2
	// NameTypeSRVHST is the service with host name type (NT-SRV-HST).
	NameTypeSRVHST = 3
	// NameTypeEnterprise is the enterprise name type (NT-ENTERPRISE).
	NameTypeEnterprise = 10
)

// Encryption type constants (RFC 3961, RFC 3962, RFC 4757).
const (
	// ETypeRC4HMAC is the RC4-HMAC encryption type (etype 23), per RFC 4757.
	ETypeRC4HMAC = 23
	// ETypeAES128CTSHMACSHA196 is AES-128-CTS-HMAC-SHA1-96 (etype 17), per RFC 3962.
	ETypeAES128CTSHMACSHA196 = 17
	// ETypeAES256CTSHMACSHA196 is AES-256-CTS-HMAC-SHA1-96 (etype 18), per RFC 3962.
	ETypeAES256CTSHMACSHA196 = 18
)

// Pre-authentication data type constants (RFC 4120 Section 7.5.2).
const (
	// PATGSReq is the TGS-REQ pre-auth type (PA-TGS-REQ).
	PATGSReq = 1
	// PAEncTimestamp is the encrypted timestamp pre-auth type (PA-ENC-TIMESTAMP).
	PAEncTimestamp = 2
	// PAETypeInfo2 is the encryption type info version 2 pre-auth type (PA-ETYPE-INFO2).
	PAETypeInfo2 = 19
	// PAPACRequest is the Microsoft PA-PAC-REQUEST pre-auth type (MS-KILE).
	// Required by Windows KDCs to avoid silent dropping of AS-REQs without pre-auth.
	PAPACRequest = 128
)

// KDC error code constants (RFC 4120 Section 7.5.9).
const (
	// ErrNone indicates no error.
	ErrNone = 0
	// ErrCPrincipalUnknown is KDC_ERR_C_PRINCIPAL_UNKNOWN: client not found in database.
	ErrCPrincipalUnknown = 6
	// ErrKDCUnavailable is KRB_ERR_GENERIC: KDC unavailable.
	ErrKDCUnavailable = 13
	// ErrPreauthRequired is KDC_ERR_PREAUTH_REQUIRED: pre-authentication required.
	ErrPreauthRequired = 25
)

// AP options bit position constants (RFC 4120 Section 5.5.1).
const (
	// APOptionUseSessionKey requests use of session key instead of service key.
	APOptionUseSessionKey = 1
	// APOptionMutualAuth requests mutual authentication.
	APOptionMutualAuth = 2
)

// Ticket flag bit position constants (RFC 4120 Section 2.1).
const (
	// TicketFlagForwardable marks the ticket as forwardable.
	TicketFlagForwardable = 1
	// TicketFlagForwarded marks the ticket as forwarded.
	TicketFlagForwarded = 2
	// TicketFlagProxiable marks the ticket as proxiable.
	TicketFlagProxiable = 3
	// TicketFlagProxy marks the ticket as a proxy ticket.
	TicketFlagProxy = 4
	// TicketFlagPreAuthent marks the ticket as pre-authenticated.
	TicketFlagPreAuthent = 6
	// TicketFlagInitial marks the ticket as an initial ticket.
	TicketFlagInitial = 7
	// TicketFlagRenewable marks the ticket as renewable.
	TicketFlagRenewable = 8
)
