package mspan

import (
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// PrintAsyncNotificationType is a notification type identifier ([MS-PAN] 2.2.1): a
// typedef of GUID in the IDL, so it is carried on the wire as a 16-octet GUID
// (Data1/Data2/Data3 + Data4[8], 4-octet aligned).
//
// It is modeled on msdtyp.GUID rather than windows/guid.GUID because the latter's trailing
// uint64 does not marshal to the required 16 octets under NDR.
type PrintAsyncNotificationType msdtyp.GUID

// GUID returns the notification type as a windows/guid.GUID for display and comparison.
func (t PrintAsyncNotificationType) GUID() guid.GUID { return msdtyp.GUID(t).GUID() }

// String returns the canonical brace-less GUID string form of the notification type.
func (t PrintAsyncNotificationType) String() string { return msdtyp.GUID(t).String() }
