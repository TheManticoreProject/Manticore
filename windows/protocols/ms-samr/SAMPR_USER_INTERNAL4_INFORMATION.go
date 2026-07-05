package mssamr

// SAMPR_USER_INTERNAL4_INFORMATION carries the full user attribute set plus an
// encrypted password ([MS-SAMR] 2.2.6.25).
type SAMPR_USER_INTERNAL4_INFORMATION struct {
	I1           SAMPR_USER_ALL_INFORMATION
	UserPassword SAMPR_ENCRYPTED_USER_PASSWORD
}
