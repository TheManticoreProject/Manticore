package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// RCM_REMOTEADDRESS_IPV4 is the AF_INET (sin_family == 2) arm of RCM_REMOTEADDRESS.
type RCM_REMOTEADDRESS_IPV4 struct {
	SinPort uint16
	InAddr  ndr.DWORD
	SinZero [8]uint8
}

// RCM_REMOTEADDRESS_IPV6 is the AF_INET6 (sin_family == 23) arm of RCM_REMOTEADDRESS.
type RCM_REMOTEADDRESS_IPV6 struct {
	Sin6Port     uint16
	Sin6Flowinfo ndr.DWORD
	Sin6Addr     [8]uint16
	Sin6ScopeId  ndr.DWORD
}

// RCM_REMOTEADDRESS carries the remote network address of a session returned by
// RpcGetRemoteAddress ([MS-TSTS] 2.2.2.1, rcmpublic.idl _RCM_REMOTEADDRESS). It is an NDR
// encapsulated union discriminated by the 16-bit sin_family; the discriminant is
// transmitted inline ahead of the selected arm.
type RCM_REMOTEADDRESS struct {
	SinFamily uint16                 `ndr:"switch"`
	Ipv4      RCM_REMOTEADDRESS_IPV4 `ndr:"case=2"`
	Ipv6      RCM_REMOTEADDRESS_IPV6 `ndr:"case=23"`
}
