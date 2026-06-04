package structures

// SAMPR_USER_INTERNAL8_INFORMATION carries the full user attribute set plus an
// AES-encrypted password ([MS-SAMR] 2.2.6.31).
type SAMPR_USER_INTERNAL8_INFORMATION struct {
	I1           SAMPR_USER_ALL_INFORMATION
	UserPassword SAMPR_ENCRYPTED_PASSWORD_AES
}
