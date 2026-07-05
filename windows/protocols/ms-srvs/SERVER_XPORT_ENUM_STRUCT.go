package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_XPORT_ENUM_STRUCT is the server transport enumeration structure
// ([MS-SRVS] 2.2.4.103). Level selects the arm of the embedded
// SERVER_XPORT_ENUM_UNION (which carries its own switch discriminant).
type SERVER_XPORT_ENUM_STRUCT struct {
	Level     ndr.DWORD
	XportInfo SERVER_XPORT_ENUM_UNION
}
