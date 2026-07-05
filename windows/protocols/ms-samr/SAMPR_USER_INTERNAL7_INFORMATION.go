package mssamr

// SAMPR_USER_INTERNAL7_INFORMATION carries an AES-encrypted password and an
// expiration flag ([MS-SAMR] 2.2.6.30).
type SAMPR_USER_INTERNAL7_INFORMATION struct {
	UserPassword    SAMPR_ENCRYPTED_PASSWORD_AES
	PasswordExpired bool
}
