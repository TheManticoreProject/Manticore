package structures

// LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION_INTERNAL communicates identification, POSIX
// offset, and internal (blob) authentication information for a trusted domain ([MS-LSAD]
// 2.2.7.13).
type LSAPR_TRUSTED_DOMAIN_FULL_INFORMATION_INTERNAL struct {
	Information     LSAPR_TRUSTED_DOMAIN_INFORMATION_EX
	PosixOffset     TRUSTED_POSIX_OFFSET_INFO
	AuthInformation LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL
}
