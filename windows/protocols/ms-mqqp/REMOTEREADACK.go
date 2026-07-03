package msmqqp

// REMOTEREADACK is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-MQQP]).
type REMOTEREADACK uint16

const (
	RR_UNKNOWN REMOTEREADACK = 0
	RR_NACK    REMOTEREADACK = 1
	RR_ACK     REMOTEREADACK = 2
)
