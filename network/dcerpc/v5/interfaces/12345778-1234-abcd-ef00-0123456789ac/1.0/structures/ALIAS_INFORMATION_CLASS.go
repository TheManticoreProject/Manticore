package structures

// ALIAS_INFORMATION_CLASS enumerates the alias information classes that select the arm of
// SAMPR_ALIAS_INFO_BUFFER ([MS-SAMR] 2.2.6.5). As an NDR enum it is transmitted as a
// 16-bit unsigned value ([C706] section 14.3.6).
type ALIAS_INFORMATION_CLASS uint16

const (
	AliasGeneralInformation      ALIAS_INFORMATION_CLASS = 1
	AliasNameInformation         ALIAS_INFORMATION_CLASS = 2
	AliasAdminCommentInformation ALIAS_INFORMATION_CLASS = 3
)
