package structures

// DOMAIN_LOGOFF_INFORMATION holds the forced-logoff time for a domain
// ([MS-SAMR] 2.2.4.6). ForceLogoff is an OLD_LARGE_INTEGER defined by the base family.
type DOMAIN_LOGOFF_INFORMATION struct {
	ForceLogoff OLD_LARGE_INTEGER
}
