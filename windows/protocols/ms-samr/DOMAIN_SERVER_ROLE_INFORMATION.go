package mssamr

// DOMAIN_SERVER_ROLE_INFORMATION holds the role (backup or primary) of a server
// ([MS-SAMR] 2.2.4.4).
type DOMAIN_SERVER_ROLE_INFORMATION struct {
	DomainServerRole DOMAIN_SERVER_ROLE `ndr:"enum"`
}
