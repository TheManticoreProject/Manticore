package mswkst

// NETSETUP_JOIN_STATUS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-WKST]).
type NETSETUP_JOIN_STATUS uint16

const (
	NetSetupUnknownStatus NETSETUP_JOIN_STATUS = 0
	NetSetupUnjoined      NETSETUP_JOIN_STATUS = 1
	NetSetupWorkgroupName NETSETUP_JOIN_STATUS = 2
	NetSetupDomainName    NETSETUP_JOIN_STATUS = 3
)
