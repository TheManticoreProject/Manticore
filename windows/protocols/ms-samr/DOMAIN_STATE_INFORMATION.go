package mssamr

// DOMAIN_STATE_INFORMATION holds the enabled/disabled state of a domain
// ([MS-SAMR] 2.2.4.2).
type DOMAIN_STATE_INFORMATION struct {
	DomainServerState DOMAIN_SERVER_ENABLE_STATE `ndr:"enum"`
}
