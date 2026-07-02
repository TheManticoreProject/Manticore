package msfasp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FW_PROFILE_CONFIG_VALUE is the FW_PROFILE_CONFIG-discriminated union returned/accepted
// by the per-profile configuration methods ([MS-FASP] 2.2.62). The selected arm depends
// on the configuration option: FW_PROFILE_CONFIG_LOG_FILE_PATH (9) carries a path string,
// FW_PROFILE_CONFIG_DISABLED_INTERFACES (15) carries an interface LUID list, and every
// DWORD-valued option carries a pointer to its value. FW_PROFILE_CONFIG_INVALID (0),
// FW_PROFILE_CONFIG_MAX (19) and any unlisted option select the empty ([default]) arm.
// The inline discriminant is transmitted per [C706] 14.3.8.
type FW_PROFILE_CONFIG_VALUE struct {
	Tag                                              FW_PROFILE_CONFIG   `ndr:"switch,enum"`
	WszStr                                           *ndr.WSTR           `ndr:"case=9,unique"`
	PDisabledInterfaces                              *FW_INTERFACE_LUIDS `ndr:"case=15,unique"`
	PdwEnableFW                                      *ndr.DWORD          `ndr:"case=1,unique"`
	PdwDisableStealthMode                            *ndr.DWORD          `ndr:"case=2,unique"`
	PdwShielded                                      *ndr.DWORD          `ndr:"case=3,unique"`
	PdwDisableUnicastResponses                       *ndr.DWORD          `ndr:"case=4,unique"`
	PdwLogDroppedPackets                             *ndr.DWORD          `ndr:"case=5,unique"`
	PdwLogSuccessConnections                         *ndr.DWORD          `ndr:"case=6,unique"`
	PdwLogIgnoredRules                               *ndr.DWORD          `ndr:"case=7,unique"`
	PdwLogMaxFileSize                                *ndr.DWORD          `ndr:"case=8,unique"`
	PdwDisableInboundNotifications                   *ndr.DWORD          `ndr:"case=10,unique"`
	PdwAuthAppsAllowUserPrefMerge                    *ndr.DWORD          `ndr:"case=11,unique"`
	PdwGlobalPortsAllowUserPrefMerge                 *ndr.DWORD          `ndr:"case=12,unique"`
	PdwAllowLocalPolicyMerge                         *ndr.DWORD          `ndr:"case=13,unique"`
	PdwAllowLocalIpsecPolicyMerge                    *ndr.DWORD          `ndr:"case=14,unique"`
	PdwDefaultOutboundAction                         *ndr.DWORD          `ndr:"case=16,unique"`
	PdwDefaultInboundAction                          *ndr.DWORD          `ndr:"case=17,unique"`
	PdwDisableStealthModeIpsecSecuredPacketExemption *ndr.DWORD          `ndr:"case=18,unique"`
}
