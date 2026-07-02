// Package msmqmq holds the [MS-MQMQ] wire structures shared by the Message Queuing
// protocols. Only the subset required by [MS-MQDS] (the PROPVARIANT property variant,
// its counted-array arms, and the PROPID/VARTYPE aliases) is modelled here; other
// [MS-MQMQ] types are added as the protocols that need them are implemented.
package msmqmq

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// PROPID is a property identifier ([MS-MQMQ] 2.2.10.1): typedef unsigned long PROPID.
type PROPID = ndr.DWORD

// VARTYPE is the 2-octet discriminant of a PROPVARIANT ([MS-MQMQ] 2.2.12): typedef
// unsigned short VARTYPE. Its values come from the VARENUM enumeration below.
type VARTYPE = uint16

// VARENUM values used as the PROPVARIANT discriminant ([MS-MQMQ] 2.2.12.1). The VT_VECTOR
// flag is OR'd with a base type to select a counted-array (CA*) arm.
const (
	VT_EMPTY   VARTYPE = 0
	VT_NULL    VARTYPE = 1
	VT_I2      VARTYPE = 2
	VT_I4      VARTYPE = 3
	VT_BOOL    VARTYPE = 11
	VT_VARIANT VARTYPE = 12
	VT_I1      VARTYPE = 16
	VT_UI1     VARTYPE = 17
	VT_UI2     VARTYPE = 18
	VT_UI4     VARTYPE = 19
	VT_I8      VARTYPE = 20
	VT_UI8     VARTYPE = 21
	VT_LPWSTR  VARTYPE = 31
	VT_BLOB    VARTYPE = 65
	VT_CLSID   VARTYPE = 72

	VT_VECTOR VARTYPE = 0x1000
)

// Combined VT_VECTOR|<base> discriminants, precomputed for use as PROPVARIANT union arm
// case values (the NDR union arm tags below reference the same numeric values).
const (
	VT_VECTOR_UI1     VARTYPE = VT_VECTOR | VT_UI1     // 0x1011
	VT_VECTOR_UI2     VARTYPE = VT_VECTOR | VT_UI2     // 0x1012
	VT_VECTOR_I4      VARTYPE = VT_VECTOR | VT_I4      // 0x1003
	VT_VECTOR_UI4     VARTYPE = VT_VECTOR | VT_UI4     // 0x1013
	VT_VECTOR_UI8     VARTYPE = VT_VECTOR | VT_UI8     // 0x1015
	VT_VECTOR_CLSID   VARTYPE = VT_VECTOR | VT_CLSID   // 0x1048
	VT_VECTOR_LPWSTR  VARTYPE = VT_VECTOR | VT_LPWSTR  // 0x101F
	VT_VECTOR_VARIANT VARTYPE = VT_VECTOR | VT_VARIANT // 0x100C
)
