package msfasp

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// PWSTR is a pointer to a null-terminated wide string (wtypes LPWSTR) ([MS-FASP]).
type PWSTR = *ndr.WSTR
