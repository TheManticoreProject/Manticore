package msnrpc

// NETLOGON_VALIDATION is a discriminated union ([MS-NRPC]); the discriminant precedes the
// selected arm ([C706] 14.3.8). Every arm is a P* pointer typedef in the IDL
// (PNETLOGON_VALIDATION_SAM_INFO, ...), so it is modeled as a [unique] pointer: the arm
// is a referent id followed by its deferred body.
type NETLOGON_VALIDATION struct {
	Tag                NETLOGON_VALIDATION_INFO_CLASS     `ndr:"switch"`
	ValidationSam      *NETLOGON_VALIDATION_SAM_INFO      `ndr:"case=2,unique"`
	ValidationSam2     *NETLOGON_VALIDATION_SAM_INFO2     `ndr:"case=3,unique"`
	ValidationGeneric2 *NETLOGON_VALIDATION_GENERIC_INFO2 `ndr:"case=5,unique"`
	ValidationSam4     *NETLOGON_VALIDATION_SAM_INFO4     `ndr:"case=6,unique"`
	ValidationTicket   *NETLOGON_VALIDATION_TICKET_LOGON  `ndr:"case=7,unique"`
}
