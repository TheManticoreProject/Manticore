package msdfsnm

// DFS_TARGET_PRIORITY_CLASS is a [v1_enum] NDR enum ([MS-DFSNM] 2.2.2.4): [v1_enum]
// forces a 4-octet transmission (MS-RPCE 2.2.5.1), and the enumeration includes the
// negative value DfsInvalidPriorityClass (-1), so it is modeled as a signed int32.
type DFS_TARGET_PRIORITY_CLASS int32

const (
	DfsInvalidPriorityClass        DFS_TARGET_PRIORITY_CLASS = -1
	DfsSiteCostNormalPriorityClass DFS_TARGET_PRIORITY_CLASS = 0
	DfsGlobalHighPriorityClass     DFS_TARGET_PRIORITY_CLASS = 1
	DfsSiteCostHighPriorityClass   DFS_TARGET_PRIORITY_CLASS = 2
	DfsSiteCostLowPriorityClass    DFS_TARGET_PRIORITY_CLASS = 3
	DfsGlobalLowPriorityClass      DFS_TARGET_PRIORITY_CLASS = 4
)
