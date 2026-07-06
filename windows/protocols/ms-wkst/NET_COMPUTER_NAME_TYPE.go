package mswkst

// NET_COMPUTER_NAME_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-WKST]).
type NET_COMPUTER_NAME_TYPE uint16

const (
	NetPrimaryComputerName    NET_COMPUTER_NAME_TYPE = 0
	NetAlternateComputerNames NET_COMPUTER_NAME_TYPE = 1
	NetAllComputerNames       NET_COMPUTER_NAME_TYPE = 2
	NetComputerNameTypeMax    NET_COMPUTER_NAME_TYPE = 3
)
