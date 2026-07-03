package msraa

// AUTHZ_CONTEXT_INFORMATION_CLASS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RAA]).
type AUTHZ_CONTEXT_INFORMATION_CLASS uint16

const (
	AuthzContextInfoUserSid        AUTHZ_CONTEXT_INFORMATION_CLASS = 1
	AuthzContextInfoGroupsSids     AUTHZ_CONTEXT_INFORMATION_CLASS = 2
	AuthzContextInfoRestrictedSids AUTHZ_CONTEXT_INFORMATION_CLASS = 3
	ReservedEnumValue4             AUTHZ_CONTEXT_INFORMATION_CLASS = 4
	ReservedEnumValue5             AUTHZ_CONTEXT_INFORMATION_CLASS = 5
	ReservedEnumValue6             AUTHZ_CONTEXT_INFORMATION_CLASS = 6
	ReservedEnumValue7             AUTHZ_CONTEXT_INFORMATION_CLASS = 7
	ReservedEnumValue8             AUTHZ_CONTEXT_INFORMATION_CLASS = 8
	ReservedEnumValue9             AUTHZ_CONTEXT_INFORMATION_CLASS = 9
	ReservedEnumValue10            AUTHZ_CONTEXT_INFORMATION_CLASS = 10
	ReservedEnumValue11            AUTHZ_CONTEXT_INFORMATION_CLASS = 11
	AuthzContextInfoDeviceSids     AUTHZ_CONTEXT_INFORMATION_CLASS = 12
	AuthzContextInfoUserClaims     AUTHZ_CONTEXT_INFORMATION_CLASS = 13
	AuthzContextInfoDeviceClaims   AUTHZ_CONTEXT_INFORMATION_CLASS = 14
	ReservedEnumValue15            AUTHZ_CONTEXT_INFORMATION_CLASS = 15
	ReservedEnumValue16            AUTHZ_CONTEXT_INFORMATION_CLASS = 16
)
