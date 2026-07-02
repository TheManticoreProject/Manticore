package msfasp

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// LPSTR is a pointer to a null-terminated ANSI string (wtypes LPSTR) ([MS-FASP]).
type LPSTR = *ndr.STR
