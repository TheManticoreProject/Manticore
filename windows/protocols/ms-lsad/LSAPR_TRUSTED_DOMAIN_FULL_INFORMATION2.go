package mslsad

// LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION2 communicates identification, POSIX offset, and
// authentication information for a trusted domain, including forest-trust data ([MS-LSAD]
// 2.2.7.12).
type LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION2 struct {
	Information     LSAPR_TRUSTED_DOMAIN_INFORMATION_EX2
	PosixOffset     TRUSTED_POSIX_OFFSET_INFO
	AuthInformation LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION
}
