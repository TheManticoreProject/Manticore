package mslsad

// LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL wraps an encrypted authentication blob
// for a trust ([MS-LSAD] 2.2.7.15). AuthBlob is embedded inline.
type LSAPR_TRUSTED_DOMAIN_AUTH_INFORMATION_INTERNAL struct {
	AuthBlob LSAPR_TRUSTED_DOMAIN_AUTH_BLOB
}
