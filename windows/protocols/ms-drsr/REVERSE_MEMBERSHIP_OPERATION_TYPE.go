package msdrsr

// REVERSE_MEMBERSHIP_OPERATION_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DRSR]).
type REVERSE_MEMBERSHIP_OPERATION_TYPE uint16

const (
	RevMembGetGroupsForUser          REVERSE_MEMBERSHIP_OPERATION_TYPE = 1
	RevMembGetAliasMembership        REVERSE_MEMBERSHIP_OPERATION_TYPE = 2
	RevMembGetAccountGroups          REVERSE_MEMBERSHIP_OPERATION_TYPE = 3
	RevMembGetResourceGroups         REVERSE_MEMBERSHIP_OPERATION_TYPE = 4
	RevMembGetUniversalGroups        REVERSE_MEMBERSHIP_OPERATION_TYPE = 5
	GroupMembersTransitive           REVERSE_MEMBERSHIP_OPERATION_TYPE = 6
	RevMembGlobalGroupsNonTransitive REVERSE_MEMBERSHIP_OPERATION_TYPE = 7
)
