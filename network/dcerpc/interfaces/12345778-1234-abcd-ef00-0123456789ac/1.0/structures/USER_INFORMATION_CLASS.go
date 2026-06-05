package structures

// USER_INFORMATION_CLASS enumerates the user information levels used by the
// SAMR user query/set methods ([MS-SAMR] 2.2.6.28). NDR enums are 16-bit, so
// this is modeled as a uint16. The enumeration has gaps (15, 19, 22, 27, 29,
// 30) that have no defined value.
type USER_INFORMATION_CLASS uint16

const (
	UserGeneralInformation      USER_INFORMATION_CLASS = 1
	UserPreferencesInformation  USER_INFORMATION_CLASS = 2
	UserLogonInformation        USER_INFORMATION_CLASS = 3
	UserLogonHoursInformation   USER_INFORMATION_CLASS = 4
	UserAccountInformation      USER_INFORMATION_CLASS = 5
	UserNameInformation         USER_INFORMATION_CLASS = 6
	UserAccountNameInformation  USER_INFORMATION_CLASS = 7
	UserFullNameInformation     USER_INFORMATION_CLASS = 8
	UserPrimaryGroupInformation USER_INFORMATION_CLASS = 9
	UserHomeInformation         USER_INFORMATION_CLASS = 10
	UserScriptInformation       USER_INFORMATION_CLASS = 11
	UserProfileInformation      USER_INFORMATION_CLASS = 12
	UserAdminCommentInformation USER_INFORMATION_CLASS = 13
	UserWorkStationsInformation USER_INFORMATION_CLASS = 14
	UserControlInformation      USER_INFORMATION_CLASS = 16
	UserExpiresInformation      USER_INFORMATION_CLASS = 17
	UserInternal1Information    USER_INFORMATION_CLASS = 18
	UserParametersInformation   USER_INFORMATION_CLASS = 20
	UserAllInformation          USER_INFORMATION_CLASS = 21
	UserInternal4Information    USER_INFORMATION_CLASS = 23
	UserInternal5Information    USER_INFORMATION_CLASS = 24
	UserInternal4InformationNew USER_INFORMATION_CLASS = 25
	UserInternal5InformationNew USER_INFORMATION_CLASS = 26
	UserInternal7Information    USER_INFORMATION_CLASS = 31
	UserInternal8Information    USER_INFORMATION_CLASS = 32
)
