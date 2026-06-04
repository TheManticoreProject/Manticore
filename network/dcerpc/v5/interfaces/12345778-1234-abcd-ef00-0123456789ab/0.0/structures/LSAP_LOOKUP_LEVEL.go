package structures

// LSAP_LOOKUP_LEVEL specifies the scope of a SID/name lookup translation request
// ([MS-LSAT] 2.2.16). As an NDR enum it is transmitted as a 16-bit unsigned value
// ([C706] section 14.3.6).
type LSAP_LOOKUP_LEVEL uint16

const (
	LsapLookupWksta             LSAP_LOOKUP_LEVEL = 1
	LsapLookupPDC               LSAP_LOOKUP_LEVEL = 2
	LsapLookupTDL               LSAP_LOOKUP_LEVEL = 3
	LsapLookupGC                LSAP_LOOKUP_LEVEL = 4
	LsapLookupXForestReferral   LSAP_LOOKUP_LEVEL = 5
	LsapLookupXForestResolve    LSAP_LOOKUP_LEVEL = 6
	LsapLookupRODCReferralToFullDC LSAP_LOOKUP_LEVEL = 7
)
