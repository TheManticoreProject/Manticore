package msraa

// AUTHZR_CONTEXT_INFORMATION carries one piece of client-context information returned by
// AuthzGetInformationFromContext ([MS-RAA] 2.2.3.7). ValueType selects the arm of
// ContextInfoUnion via switch_is(ValueType); being a non-encapsulated union, the
// discriminant is transmitted twice — once as this ValueType field and once inline at the
// head of the union ([C706] 14.3.8).
type AUTHZR_CONTEXT_INFORMATION struct {
	ValueType        uint16
	ContextInfoUnion AUTHZR_CONTEXT_INFORMATION_UNION
}

// AUTHZR_CONTEXT_INFORMATION_UNION is the switch_is(ValueType) union of
// AUTHZR_CONTEXT_INFORMATION ([MS-RAA] 2.2.3.7). The discriminant is the enclosing USHORT
// ValueType, so Tag is a 16-bit scalar transmitted inline ahead of the selected arm
// ([C706] 14.3.8). Every arm is a [unique] pointer. The IDL maps several
// AUTHZ_CONTEXT_INFORMATION_CLASS values to one arm — token groups for 0x2 (GroupsSids),
// 0x3 (RestrictedSids), and 0xC (DeviceSids); token claims for 0xD (UserClaims) and 0xE
// (DeviceClaims) — and the declarative union takes one case value per field, so those arms
// are repeated per case value.
type AUTHZR_CONTEXT_INFORMATION_UNION struct {
	Tag                 uint16                                  `ndr:"switch"`
	PTokenUser          *AUTHZR_TOKEN_USER                      `ndr:"case=1,unique"`
	PTokenGroups        *AUTHZR_TOKEN_GROUPS                    `ndr:"case=2,unique"`
	PTokenRestrictedSid *AUTHZR_TOKEN_GROUPS                    `ndr:"case=3,unique"`
	PTokenDeviceGroups  *AUTHZR_TOKEN_GROUPS                    `ndr:"case=12,unique"`
	PTokenUserClaims    *AUTHZR_SECURITY_ATTRIBUTES_INFORMATION `ndr:"case=13,unique"`
	PTokenDeviceClaims  *AUTHZR_SECURITY_ATTRIBUTES_INFORMATION `ndr:"case=14,unique"`
}
