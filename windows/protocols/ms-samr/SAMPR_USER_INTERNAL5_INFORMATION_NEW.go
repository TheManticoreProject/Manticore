package mssamr

// SAMPR_USER_INTERNAL5_INFORMATION_NEW carries an encrypted password with salt
// and an expiration flag ([MS-SAMR] 2.2.6.28).
type SAMPR_USER_INTERNAL5_INFORMATION_NEW struct {
	UserPassword    SAMPR_ENCRYPTED_USER_PASSWORD_NEW
	PasswordExpired uint8
}
