package structures

// SAMPR_DISPLAY_INFO_BUFFER is the discriminated union of domain display information
// classes ([MS-SAMR] 2.2.8.12). The discriminant Tag is a DOMAIN_DISPLAY_INFORMATION;
// the wire form is the discriminant followed by the single selected arm ([C706] section
// 14.3.8). The numeric case values follow the DOMAIN_DISPLAY_INFORMATION enum.
type SAMPR_DISPLAY_INFO_BUFFER struct {
	Tag                 DOMAIN_DISPLAY_INFORMATION            `ndr:"switch,enum"`
	UserInformation     SAMPR_DOMAIN_DISPLAY_USER_BUFFER      `ndr:"case=1"`
	MachineInformation  SAMPR_DOMAIN_DISPLAY_MACHINE_BUFFER   `ndr:"case=2"`
	GroupInformation    SAMPR_DOMAIN_DISPLAY_GROUP_BUFFER     `ndr:"case=3"`
	OemUserInformation  SAMPR_DOMAIN_DISPLAY_OEM_USER_BUFFER  `ndr:"case=4"`
	OemGroupInformation SAMPR_DOMAIN_DISPLAY_OEM_GROUP_BUFFER `ndr:"case=5"`
}
