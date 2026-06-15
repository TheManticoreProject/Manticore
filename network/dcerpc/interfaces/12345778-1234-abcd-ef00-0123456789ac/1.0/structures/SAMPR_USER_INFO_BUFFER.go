package structures

// SAMPR_USER_INFO_BUFFER is the union of all user information structures,
// discriminated by a USER_INFORMATION_CLASS ([MS-SAMR] 2.2.6.29). Tag carries
// the discriminant inline; exactly one arm is valid per the selected case.
type SAMPR_USER_INFO_BUFFER struct {
	Tag USER_INFORMATION_CLASS `ndr:"switch,enum"`

	General      SAMPR_USER_GENERAL_INFORMATION       `ndr:"case=1"`
	Preferences  SAMPR_USER_PREFERENCES_INFORMATION   `ndr:"case=2"`
	Logon        SAMPR_USER_LOGON_INFORMATION         `ndr:"case=3"`
	LogonHours   SAMPR_USER_LOGON_HOURS_INFORMATION   `ndr:"case=4"`
	Account      SAMPR_USER_ACCOUNT_INFORMATION       `ndr:"case=5"`
	Name         SAMPR_USER_NAME_INFORMATION          `ndr:"case=6"`
	AccountName  SAMPR_USER_A_NAME_INFORMATION        `ndr:"case=7"`
	FullName     SAMPR_USER_F_NAME_INFORMATION        `ndr:"case=8"`
	PrimaryGroup USER_PRIMARY_GROUP_INFORMATION       `ndr:"case=9"`
	Home         SAMPR_USER_HOME_INFORMATION          `ndr:"case=10"`
	Script       SAMPR_USER_SCRIPT_INFORMATION        `ndr:"case=11"`
	Profile      SAMPR_USER_PROFILE_INFORMATION       `ndr:"case=12"`
	AdminComment SAMPR_USER_ADMIN_COMMENT_INFORMATION `ndr:"case=13"`
	WorkStations SAMPR_USER_WORKSTATIONS_INFORMATION  `ndr:"case=14"`
	Control      USER_CONTROL_INFORMATION             `ndr:"case=16"`
	Expires      USER_EXPIRES_INFORMATION             `ndr:"case=17"`
	Internal1    SAMPR_USER_INTERNAL1_INFORMATION     `ndr:"case=18"`
	Parameters   SAMPR_USER_PARAMETERS_INFORMATION    `ndr:"case=20"`
	All          SAMPR_USER_ALL_INFORMATION           `ndr:"case=21"`
	Internal4    SAMPR_USER_INTERNAL4_INFORMATION     `ndr:"case=23"`
	Internal5    SAMPR_USER_INTERNAL5_INFORMATION     `ndr:"case=24"`
	Internal4New SAMPR_USER_INTERNAL4_INFORMATION_NEW `ndr:"case=25"`
	Internal5New SAMPR_USER_INTERNAL5_INFORMATION_NEW `ndr:"case=26"`
	Internal7    SAMPR_USER_INTERNAL7_INFORMATION     `ndr:"case=31"`
	Internal8    SAMPR_USER_INTERNAL8_INFORMATION     `ndr:"case=32"`
}
