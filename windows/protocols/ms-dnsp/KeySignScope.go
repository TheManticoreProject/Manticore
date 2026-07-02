package msdnsp

// KeySignScope is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type KeySignScope uint16

const (
	SIGN_SCOPE_DEFAULT        KeySignScope = 0
	SIGN_SCOPE_DNSKEY_ONLY    KeySignScope = 1
	SIGN_SCOPE_ALL_RECORDS    KeySignScope = 2
	SIGN_SCOPE_ADD_ONLY       KeySignScope = 3
	SIGN_SCOPE_DO_NOT_PUBLISH KeySignScope = 4
	SIGN_SCOPE_REVOKED        KeySignScope = 5
)
