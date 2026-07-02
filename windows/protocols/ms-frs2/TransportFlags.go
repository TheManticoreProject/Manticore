package msfrs2

// TransportFlags is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FRS2]).
type TransportFlags uint16

const (
	TRANSPORT_SUPPORTS_RDC_SIMILARITY TransportFlags = 1
)
