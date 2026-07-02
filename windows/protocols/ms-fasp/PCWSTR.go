package msfasp

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// PCWSTR is a pointer to a null-terminated constant wide string (wtypes LPCWSTR) ([MS-FASP]).
type PCWSTR = *ndr.WSTR
