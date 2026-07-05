package msscmr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SC_RPC_CONFIG_INFOA ([MS-SCMR] 2.2.22) is the ANSI form of SC_RPC_CONFIG_INFOW: an info
// level plus a [switch_is(dwInfoLevel)] union. dwInfoLevel is also re-transmitted inline as
// the union discriminant ([C706] 14.3.8).
type SC_RPC_CONFIG_INFOA struct {
	DwInfoLevel ndr.DWORD
	Field       SC_RPC_CONFIG_INFOA_Field
}

// SC_RPC_CONFIG_INFOA_Field is the [switch_is(dwInfoLevel)] union of SC_RPC_CONFIG_INFOA.
// Each IDL arm is a [unique] pointer (LP*/P*), so each arm emits a referent id with its body
// deferred ([C706] 14.3.8). Tag is the inline 4-byte discriminant.
type SC_RPC_CONFIG_INFOA_Field struct {
	Tag   ndr.DWORD                             `ndr:"switch"`
	Psd   *SERVICE_DESCRIPTIONA                 `ndr:"case=1,unique"`
	Psfa  *SERVICE_FAILURE_ACTIONSA             `ndr:"case=2,unique"`
	Psda  *SERVICE_DELAYED_AUTO_START_INFO      `ndr:"case=3,unique"`
	Psfaf *SERVICE_FAILURE_ACTIONS_FLAG         `ndr:"case=4,unique"`
	Pssid *SERVICE_SID_INFO                     `ndr:"case=5,unique"`
	Psrp  *SERVICE_RPC_REQUIRED_PRIVILEGES_INFO `ndr:"case=6,unique"`
	Psps  *SERVICE_PRESHUTDOWN_INFO             `ndr:"case=7,unique"`
	Psti  *SERVICE_TRIGGER_INFO                 `ndr:"case=8,unique"`
	Pspn  *SERVICE_PREFERRED_NODE_INFO          `ndr:"case=9,unique"`
}
