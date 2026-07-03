package msraa

// AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE models a single typed value of a claim security
// attribute ([MS-RAA] 2.2.3.5). ValueType selects the arm of AttributeUnion via the
// switch_is(ValueType) discriminant; because it is a non-encapsulated union the
// discriminant is transmitted twice — once as this ValueType field and once inline at the
// head of the union ([C706] 14.3.8).
type AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE struct {
	ValueType      uint16
	AttributeUnion AUTHZR_SECURITY_ATTRIBUTE_UNION
}

// AUTHZR_SECURITY_ATTRIBUTE_UNION is the switch_is(ValueType) union of
// AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE ([MS-RAA] 2.2.3.5). The discriminant is the enclosing
// USHORT ValueType, so Tag is a 16-bit scalar, transmitted inline ahead of the selected
// arm ([C706] 14.3.8). Uint64 covers ValueType 0x2 (UINT64) and 0x6 (BOOLEAN, encoded as
// a UINT64); the IDL maps both case labels to one arm, so each gets its own field per the
// declarative-union rule (one case value per field).
type AUTHZR_SECURITY_ATTRIBUTE_UNION struct {
	Tag        uint16                                 `ndr:"switch"`
	Int64      int64                                  `ndr:"case=1"`
	Uint64     uint64                                 `ndr:"case=2"`
	Uint64Bool uint64                                 `ndr:"case=6"`
	String     AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE `ndr:"case=3"`
}
