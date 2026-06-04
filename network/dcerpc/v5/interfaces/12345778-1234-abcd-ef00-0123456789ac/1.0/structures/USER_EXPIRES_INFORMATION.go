package structures

// USER_EXPIRES_INFORMATION holds a user's account-expiration time ([MS-SAMR]
// 2.2.6.15).
type USER_EXPIRES_INFORMATION struct {
	AccountExpires OLD_LARGE_INTEGER
}
