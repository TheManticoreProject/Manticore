package structures

// LSAPR_TRUSTED_PASSWORD_INFO communicates the current and previous passwords of a trust
// object ([MS-LSAD] 2.2.7.7). Each member is a [unique] pointer to an
// LSAPR_CR_CIPHER_VALUE.
type LSAPR_TRUSTED_PASSWORD_INFO struct {
	Password    *LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
	OldPassword *LSAPR_CR_CIPHER_VALUE `ndr:"unique"`
}
