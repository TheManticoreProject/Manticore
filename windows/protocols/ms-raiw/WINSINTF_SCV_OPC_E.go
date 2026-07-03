package msraiw

// WINSINTF_SCV_OPC_E is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RAIW]).
type WINSINTF_SCV_OPC_E uint16

const (
	WINSINTF_E_SCV_GENERAL WINSINTF_SCV_OPC_E = 0
	WINSINTF_E_SCV_VERIFY  WINSINTF_SCV_OPC_E = 1
)
