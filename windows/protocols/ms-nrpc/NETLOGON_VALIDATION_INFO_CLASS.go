package msnrpc

// NETLOGON_VALIDATION_INFO_CLASS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-NRPC]).
type NETLOGON_VALIDATION_INFO_CLASS uint16

const (
	NetlogonValidationUasInfo      NETLOGON_VALIDATION_INFO_CLASS = 1
	NetlogonValidationSamInfo      NETLOGON_VALIDATION_INFO_CLASS = 2
	NetlogonValidationSamInfo2     NETLOGON_VALIDATION_INFO_CLASS = 3
	NetlogonValidationGenericInfo  NETLOGON_VALIDATION_INFO_CLASS = 4
	NetlogonValidationGenericInfo2 NETLOGON_VALIDATION_INFO_CLASS = 5
	NetlogonValidationSamInfo4     NETLOGON_VALIDATION_INFO_CLASS = 6
	NetlogonValidationTicketLogon  NETLOGON_VALIDATION_INFO_CLASS = 7
)
