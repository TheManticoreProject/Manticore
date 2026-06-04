package structures

// SAMPR_ALIAS_INFO_BUFFER is the discriminated union of alias information classes
// ([MS-SAMR] 2.2.6.6). The discriminant is an ALIAS_INFORMATION_CLASS; the wire form is
// the discriminant followed by the single selected arm ([C706] section 14.3.8).
type SAMPR_ALIAS_INFO_BUFFER struct {
	Tag          ALIAS_INFORMATION_CLASS             `ndr:"switch"`
	General      SAMPR_ALIAS_GENERAL_INFORMATION     `ndr:"case=1"`
	Name         SAMPR_ALIAS_NAME_INFORMATION        `ndr:"case=2"`
	AdminComment SAMPR_ALIAS_ADM_COMMENT_INFORMATION `ndr:"case=3"`
}
