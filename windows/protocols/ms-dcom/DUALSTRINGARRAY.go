package msdcom

// DUALSTRINGARRAY models DUALSTRINGARRAY ([MS-DCOM] 2.2.19): a marshaled set of network
// and security bindings. aStringArray is a bare inline conformant array (a member with
// [], not a pointer), so its maximum_count is transmitted in place with no referent id —
// it is tagged "conformant", never "unique". It holds wNumEntries unsigned shorts: a
// sequence of STRINGBINDING entries (network bindings) terminated by a null, followed at
// wSecurityOffset by a sequence of SECURITYBINDING entries terminated by a null.
type DUALSTRINGARRAY struct {
	WNumEntries     uint16
	WSecurityOffset uint16
	AStringArray    []uint16 `ndr:"conformant,size_is=WNumEntries"`
}
