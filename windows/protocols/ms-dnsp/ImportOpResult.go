package msdnsp

// ImportOpResult is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-DNSP]).
type ImportOpResult uint16

const (
	IMPORT_STATUS_NOOP            ImportOpResult = 0
	IMPORT_STATUS_SIGNING_READY   ImportOpResult = 1
	IMPORT_STATUS_UNSIGNING_READY ImportOpResult = 2
	IMPORT_STATUS_CHANGED         ImportOpResult = 3
)
