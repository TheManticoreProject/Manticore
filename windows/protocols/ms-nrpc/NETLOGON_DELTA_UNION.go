package msnrpc

// NETLOGON_DELTA_UNION is a discriminated union ([MS-NRPC]); the discriminant precedes the
// selected arm ([C706] 14.3.8). Every arm is a P* pointer typedef in the IDL, so each is a
// [unique] pointer: a referent id inline, the body deferred to the end of the enclosing
// construction (the NETLOGON_DELTA_ENUM array). The [default] arm is empty.
type NETLOGON_DELTA_UNION struct {
	Tag                   NETLOGON_DELTA_TYPE             `ndr:"switch"`
	DeltaDomain           *NETLOGON_DELTA_DOMAIN          `ndr:"case=1,unique"`
	DeltaGroup            *NETLOGON_DELTA_GROUP           `ndr:"case=2,unique"`
	DeltaRenameGroup      *NETLOGON_RENAME_GROUP          `ndr:"case=4,unique"`
	DeltaUser             *NETLOGON_DELTA_USER            `ndr:"case=5,unique"`
	DeltaRenameUser       *NETLOGON_RENAME_USER           `ndr:"case=7,unique"`
	DeltaGroupMember      *NETLOGON_DELTA_GROUP_MEMBER    `ndr:"case=8,unique"`
	DeltaAlias            *NETLOGON_DELTA_ALIAS           `ndr:"case=9,unique"`
	DeltaRenameAlias      *NETLOGON_RENAME_ALIAS          `ndr:"case=11,unique"`
	DeltaAliasMember      *NETLOGON_DELTA_ALIAS_MEMBER    `ndr:"case=12,unique"`
	DeltaPolicy           *NETLOGON_DELTA_POLICY          `ndr:"case=13,unique"`
	DeltaTDomains         *NETLOGON_DELTA_TRUSTED_DOMAINS `ndr:"case=14,unique"`
	DeltaAccounts         *NETLOGON_DELTA_ACCOUNTS        `ndr:"case=16,unique"`
	DeltaSecret           *NETLOGON_DELTA_SECRET          `ndr:"case=18,unique"`
	DeltaDeleteGroup      *NETLOGON_DELTA_DELETE_GROUP    `ndr:"case=20,unique"`
	DeltaDeleteUser       *NETLOGON_DELTA_DELETE_USER     `ndr:"case=21,unique"`
	DeltaSerialNumberSkip *NLPR_MODIFIED_COUNT            `ndr:"case=22,unique"`
}
