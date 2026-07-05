package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// TSVIPSession is the virtual-IP assignment for a session returned by RpcGetSessionIP
// ([MS-TSTS] 2.2.2.5.3, allproc.h _TSVIPSession).
type TSVIPSession struct {
	DwVersion ndr.DWORD
	SessionId ndr.DWORD
	SessionIP TSVIPAddress
}
