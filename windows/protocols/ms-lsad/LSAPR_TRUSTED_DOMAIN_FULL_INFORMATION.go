package mslsad

// LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION communicates identification, POSIX offset, and
// authentication information for a trusted domain ([MS-LSAD] 2.2.7.10).
type LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION struct {
	Information     LSAPR_TRUSTED_DOMAIN_INFORMATION_EX
	PosixOffset     TRUSTED_POSIX_OFFSET_INFO
	AuthInformation LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION
}
