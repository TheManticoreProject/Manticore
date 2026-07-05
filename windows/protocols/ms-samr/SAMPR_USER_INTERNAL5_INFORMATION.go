package mssamr

// SAMPR_USER_INTERNAL5_INFORMATION carries an encrypted password and an
// expiration flag ([MS-SAMR] 2.2.6.26).
type SAMPR_USER_INTERNAL5_INFORMATION struct {
	UserPassword    SAMPR_ENCRYPTED_USER_PASSWORD
	PasswordExpired uint8
}
