package structures

// SAMPR_USER_LOGON_HOURS_INFORMATION holds a user's allowed logon hours
// ([MS-SAMR] 2.2.6.16).
type SAMPR_USER_LOGON_HOURS_INFORMATION struct {
	LogonHours SAMPR_LOGON_HOURS
}
