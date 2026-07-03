package msraiw

// WINSINTF_CMD_E is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RAIW]).
type WINSINTF_CMD_E uint16

const (
	WINSINTF_E_ADDVERSMAP      WINSINTF_CMD_E = 0
	WINSINTF_E_CONFIG          WINSINTF_CMD_E = 1
	WINSINTF_E_STAT            WINSINTF_CMD_E = 2
	WINSINTF_E_CONFIG_ALL_MAPS WINSINTF_CMD_E = 3
)
