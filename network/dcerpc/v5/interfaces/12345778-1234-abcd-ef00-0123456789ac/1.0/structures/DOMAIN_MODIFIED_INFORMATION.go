package structures

// DOMAIN_MODIFIED_INFORMATION holds the domain modification count and creation time
// ([MS-SAMR] 2.2.4.7). Both fields are OLD_LARGE_INTEGER values defined by the base
// family.
type DOMAIN_MODIFIED_INFORMATION struct {
	DomainModifiedCount OLD_LARGE_INTEGER
	CreationTime        OLD_LARGE_INTEGER
}
