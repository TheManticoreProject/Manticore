package structures

// EBranchOfficeJobEventType is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RPRN]).
type EBranchOfficeJobEventType uint16

const (
	kInvalidJobState     EBranchOfficeJobEventType = 0
	kLogJobPrinted       EBranchOfficeJobEventType = 1
	kLogJobRendered      EBranchOfficeJobEventType = 2
	kLogJobError         EBranchOfficeJobEventType = 3
	kLogJobPipelineError EBranchOfficeJobEventType = 4
	kLogOfflineFileFull  EBranchOfficeJobEventType = 5
)
