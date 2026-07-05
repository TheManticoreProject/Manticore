package mssamr

// SAMPR_DOMAIN_INFO_BUFFER is the discriminated union of domain information classes
// ([MS-SAMR] 2.2.4.17). The discriminant Tag is a DOMAIN_INFORMATION_CLASS; the wire
// form is the discriminant followed by the single selected arm ([C706] section 14.3.8).
// The numeric case values follow the DOMAIN_INFORMATION_CLASS enum.
type SAMPR_DOMAIN_INFO_BUFFER struct {
	Tag         DOMAIN_INFORMATION_CLASS             `ndr:"switch,enum"`
	Password    DOMAIN_PASSWORD_INFORMATION          `ndr:"case=1"`
	General     SAMPR_DOMAIN_GENERAL_INFORMATION     `ndr:"case=2"`
	Logoff      DOMAIN_LOGOFF_INFORMATION            `ndr:"case=3"`
	Oem         SAMPR_DOMAIN_OEM_INFORMATION         `ndr:"case=4"`
	Name        SAMPR_DOMAIN_NAME_INFORMATION        `ndr:"case=5"`
	Replication SAMPR_DOMAIN_REPLICATION_INFORMATION `ndr:"case=6"`
	Role        DOMAIN_SERVER_ROLE_INFORMATION       `ndr:"case=7"`
	Modified    DOMAIN_MODIFIED_INFORMATION          `ndr:"case=8"`
	State       DOMAIN_STATE_INFORMATION             `ndr:"case=9"`
	General2    SAMPR_DOMAIN_GENERAL_INFORMATION2    `ndr:"case=11"`
	Lockout     SAMPR_DOMAIN_LOCKOUT_INFORMATION     `ndr:"case=12"`
	Modified2   DOMAIN_MODIFIED_INFORMATION2         `ndr:"case=13"`
}
