package mststs

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// TSVIP_SOCKADDR_IPV4 is the AF_INET (sin_family == 2) arm of TSVIP_SOCKADDR.
type TSVIP_SOCKADDR_IPV4 struct {
	SinPort uint16
	InAddr  ndr.DWORD
	SinZero [8]uint8
}

// TSVIP_SOCKADDR_IPV6 is the AF_INET6 (sin_family == 23) arm of TSVIP_SOCKADDR.
type TSVIP_SOCKADDR_IPV6 struct {
	Sin6Port     uint16
	Sin6Flowinfo ndr.DWORD
	Sin6Addr     [8]uint16
	Sin6ScopeId  ndr.DWORD
}

// TSVIP_SOCKADDR is the encapsulated NDR union carrying either an IPv4 or IPv6 socket
// address, discriminated by the 16-bit sin_family ([MS-TSTS] 2.2.2.5.1, allproc.h
// _TSVIP_SOCKADDR). The discriminant is transmitted inline ahead of the selected arm.
type TSVIP_SOCKADDR struct {
	SinFamily uint16              `ndr:"switch"`
	Ipv4      TSVIP_SOCKADDR_IPV4 `ndr:"case=2"`
	Ipv6      TSVIP_SOCKADDR_IPV6 `ndr:"case=23"`
}
