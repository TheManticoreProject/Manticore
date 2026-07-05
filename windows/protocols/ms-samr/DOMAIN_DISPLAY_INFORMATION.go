package mssamr

// DOMAIN_DISPLAY_INFORMATION enumerates the display information classes for the
// SAMPR_DISPLAY_INFO_BUFFER union ([MS-SAMR] 2.2.8.1). As an NDR enum it is transmitted
// as a 16-bit unsigned value ([C706] section 14.3.6).
type DOMAIN_DISPLAY_INFORMATION uint16

const (
	DomainDisplayUser     DOMAIN_DISPLAY_INFORMATION = 1
	DomainDisplayMachine  DOMAIN_DISPLAY_INFORMATION = 2
	DomainDisplayGroup    DOMAIN_DISPLAY_INFORMATION = 3
	DomainDisplayOemUser  DOMAIN_DISPLAY_INFORMATION = 4
	DomainDisplayOemGroup DOMAIN_DISPLAY_INFORMATION = 5
)
