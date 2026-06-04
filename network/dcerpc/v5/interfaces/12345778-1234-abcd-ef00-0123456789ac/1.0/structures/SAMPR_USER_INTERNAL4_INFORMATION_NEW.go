package structures

// SAMPR_USER_INTERNAL4_INFORMATION_NEW carries the full user attribute set plus
// an encrypted password with salt ([MS-SAMR] 2.2.6.27).
type SAMPR_USER_INTERNAL4_INFORMATION_NEW struct {
	I1           SAMPR_USER_ALL_INFORMATION
	UserPassword SAMPR_ENCRYPTED_USER_PASSWORD_NEW
}
