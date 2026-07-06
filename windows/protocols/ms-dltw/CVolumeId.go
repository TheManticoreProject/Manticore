package msdltw

import (
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// CVolumeId is a VolumeID ([MS-DLTW] 2.2.3): the identifier assigned to a volume when
// link tracking begins on it, carried on the wire as a single 16-octet GUID (the IDL
// field _volume).
//
// The GUID is modeled on msdtyp.GUID rather than windows/guid.GUID because the latter's
// trailing uint64 does not marshal to the required 16 octets under NDR. Use Volume.GUID()
// / msdtyp.NewGUID to convert to and from windows/guid.GUID.
type CVolumeId struct {
	Volume msdtyp.GUID
}

// GUID returns the VolumeID as a windows/guid.GUID for display and comparison.
func (v CVolumeId) GUID() guid.GUID { return v.Volume.GUID() }
