package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// STAT_SERVER_0 contains statistical information about the server ([MS-SRVS]
// 2.2.4.39). All fields are DWORD; the 64-bit byte counters are split into
// _low and _high halves.
type STAT_SERVER_0 struct {
	Sts0Start         ndr.DWORD
	Sts0Fopens        ndr.DWORD
	Sts0Devopens      ndr.DWORD
	Sts0Jobsqueued    ndr.DWORD
	Sts0Sopens        ndr.DWORD
	Sts0Stimedout     ndr.DWORD
	Sts0Serrorout     ndr.DWORD
	Sts0Pwerrors      ndr.DWORD
	Sts0Permerrors    ndr.DWORD
	Sts0Syserrors     ndr.DWORD
	Sts0BytessentLow  ndr.DWORD
	Sts0BytessentHigh ndr.DWORD
	Sts0BytesrcvdLow  ndr.DWORD
	Sts0BytesrcvdHigh ndr.DWORD
	Sts0Avresponse    ndr.DWORD
	Sts0Reqbufneed    ndr.DWORD
	Sts0Bigbufneed    ndr.DWORD
}
