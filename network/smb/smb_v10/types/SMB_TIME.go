package types

import (
	"time"

	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

type SMB_TIME = data_structures.FILETIME

// NewSMB_TIMEFromTime creates a new SMB_TIME from a time.Time
//
// Parameters:
// - t: The time.Time to create the SMB_TIME from
//
// Returns:
// - The new SMB_TIME
func NewSMB_TIMEFromTime(t time.Time) *SMB_TIME {
	return data_structures.NewFILETIMEFromTime(t)
}
