package structures

// DOMAIN_SERVER_ROLE indicates the role of a server: backup (replica) or primary
// ([MS-SAMR] 2.2.4.3). As an NDR enum it is transmitted as a 16-bit unsigned value
// ([C706] section 14.3.6).
type DOMAIN_SERVER_ROLE uint16

const (
	DomainServerRoleBackup  DOMAIN_SERVER_ROLE = 2
	DomainServerRolePrimary DOMAIN_SERVER_ROLE = 3
)
