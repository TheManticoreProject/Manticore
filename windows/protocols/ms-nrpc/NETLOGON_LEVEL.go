package msnrpc

// NETLOGON_LEVEL is a discriminated union ([MS-NRPC]); the discriminant precedes the
// selected arm ([C706] 14.3.8). Every arm is a P* pointer typedef in the IDL
// (PNETLOGON_INTERACTIVE_INFO, ...), so it is modeled as a [unique] pointer: the arm
// is a referent id followed by its deferred body. Set Tag and the matching arm.
type NETLOGON_LEVEL struct {
	Tag                        NETLOGON_LOGON_INFO_CLASS   `ndr:"switch"`
	LogonInteractive           *NETLOGON_INTERACTIVE_INFO  `ndr:"case=1,unique"`
	LogonInteractiveTransitive *NETLOGON_INTERACTIVE_INFO  `ndr:"case=5,unique"`
	LogonService               *NETLOGON_SERVICE_INFO      `ndr:"case=3,unique"`
	LogonServiceTransitive     *NETLOGON_SERVICE_INFO      `ndr:"case=7,unique"`
	LogonNetwork               *NETLOGON_NETWORK_INFO      `ndr:"case=2,unique"`
	LogonNetworkTransitive     *NETLOGON_NETWORK_INFO      `ndr:"case=6,unique"`
	LogonGeneric               *NETLOGON_GENERIC_INFO      `ndr:"case=4,unique"`
	LogonTicket                *NETLOGON_TICKET_LOGON_INFO `ndr:"case=8,unique"`
}
