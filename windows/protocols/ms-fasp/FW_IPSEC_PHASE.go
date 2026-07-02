package msfasp

// FW_IPSEC_PHASE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_IPSEC_PHASE uint16

const (
	FW_IPSEC_PHASE_INVALID FW_IPSEC_PHASE = 0
	FW_IPSEC_PHASE_1       FW_IPSEC_PHASE = 1
	FW_IPSEC_PHASE_2       FW_IPSEC_PHASE = 2
	FW_IPSEC_PHASE_MAX     FW_IPSEC_PHASE = 3
)
