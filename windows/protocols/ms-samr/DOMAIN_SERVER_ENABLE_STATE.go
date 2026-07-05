package mssamr

// DOMAIN_SERVER_ENABLE_STATE indicates whether a server is enabled or disabled
// ([MS-SAMR] 2.2.4.1). As an NDR enum it is transmitted as a 16-bit unsigned value
// ([C706] section 14.3.6).
type DOMAIN_SERVER_ENABLE_STATE uint16

const (
	DomainServerEnabled  DOMAIN_SERVER_ENABLE_STATE = 1
	DomainServerDisabled DOMAIN_SERVER_ENABLE_STATE = 2
)
