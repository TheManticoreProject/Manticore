package structures

// SAMPR_GROUP_INFO_BUFFER is the discriminated union of group information classes
// ([MS-SAMR] 2.2.5.9). The discriminant is a GROUP_INFORMATION_CLASS; the wire form is the
// discriminant followed by the single selected arm ([C706] section 14.3.8).
//
// The DoNotUse arm (case GroupReplicationInformation=5) reuses the
// SAMPR_GROUP_GENERAL_INFORMATION type, as specified in the IDL.
type SAMPR_GROUP_INFO_BUFFER struct {
	Tag          GROUP_INFORMATION_CLASS             `ndr:"switch"`
	General      SAMPR_GROUP_GENERAL_INFORMATION     `ndr:"case=1"`
	Name         SAMPR_GROUP_NAME_INFORMATION        `ndr:"case=2"`
	Attribute    GROUP_ATTRIBUTE_INFORMATION         `ndr:"case=3"`
	AdminComment SAMPR_GROUP_ADM_COMMENT_INFORMATION `ndr:"case=4"`
	DoNotUse     SAMPR_GROUP_GENERAL_INFORMATION     `ndr:"case=5"`
}
