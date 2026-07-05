package mssamr

// DOMAIN_MODIFIED_INFORMATION2 extends DOMAIN_MODIFIED_INFORMATION with the modified
// count at the last promotion ([MS-SAMR] 2.2.4.8). All fields are OLD_LARGE_INTEGER
// values defined by the base family.
type DOMAIN_MODIFIED_INFORMATION2 struct {
	DomainModifiedCount          OLD_LARGE_INTEGER
	CreationTime                 OLD_LARGE_INTEGER
	ModifiedCountAtLastPromotion OLD_LARGE_INTEGER
}
