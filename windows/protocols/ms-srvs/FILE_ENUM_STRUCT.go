package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_ENUM_STRUCT is the file enumeration structure ([MS-SRVS] 2.2.4.11). Level
// selects the arm of the embedded FILE_ENUM_UNION (which carries its own switch
// discriminant).
type FILE_ENUM_STRUCT struct {
	Level    ndr.DWORD
	FileInfo FILE_ENUM_UNION
}
