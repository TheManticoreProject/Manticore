package types

import (
	"time"

	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

type SMB_TIME = msdtyp.FILETIME

// NewSMB_TIMEFromTime creates a new SMB_TIME from a time.Time
//
// Parameters:
// - t: The time.Time to create the SMB_TIME from
//
// Returns:
// - The new SMB_TIME
func NewSMB_TIMEFromTime(t time.Time) *SMB_TIME {
	return msdtyp.NewFILETIMEFromTime(t)
}
