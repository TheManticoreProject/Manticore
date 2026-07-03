package msnrpc

// NETLOGON_SECURE_CHANNEL_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-NRPC]).
type NETLOGON_SECURE_CHANNEL_TYPE uint16

const (
	NullSecureChannel             NETLOGON_SECURE_CHANNEL_TYPE = 0
	MsvApSecureChannel            NETLOGON_SECURE_CHANNEL_TYPE = 1
	WorkstationSecureChannel      NETLOGON_SECURE_CHANNEL_TYPE = 2
	TrustedDnsDomainSecureChannel NETLOGON_SECURE_CHANNEL_TYPE = 3
	TrustedDomainSecureChannel    NETLOGON_SECURE_CHANNEL_TYPE = 4
	UasServerSecureChannel        NETLOGON_SECURE_CHANNEL_TYPE = 5
	ServerSecureChannel           NETLOGON_SECURE_CHANNEL_TYPE = 6
	CdcServerSecureChannel        NETLOGON_SECURE_CHANNEL_TYPE = 7
)
