package mssamr

// GROUP_INFORMATION_CLASS enumerates the group information classes that select the arm of
// SAMPR_GROUP_INFO_BUFFER ([MS-SAMR] 2.2.5.8). As an NDR enum it is transmitted as a
// 16-bit unsigned value ([C706] section 14.3.6).
type GROUP_INFORMATION_CLASS uint16

const (
	GroupGeneralInformation      GROUP_INFORMATION_CLASS = 1
	GroupNameInformation         GROUP_INFORMATION_CLASS = 2
	GroupAttributeInformation    GROUP_INFORMATION_CLASS = 3
	GroupAdminCommentInformation GROUP_INFORMATION_CLASS = 4
	GroupReplicationInformation  GROUP_INFORMATION_CLASS = 5
)
