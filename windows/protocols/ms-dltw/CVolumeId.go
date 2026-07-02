package msdltw

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// CVolumeId is a VolumeID ([MS-DLTW] 2.2.3): the identifier assigned to a volume when
// link tracking begins on it, carried on the wire as a single 16-octet GUID (the IDL
// field _volume).
//
// The GUID is modeled on dtyp.GUID rather than windows/guid.GUID because the latter's
// trailing uint64 does not marshal to the required 16 octets under NDR. Use Volume.GUID()
// / dtyp.NewGUID to convert to and from windows/guid.GUID.
type CVolumeId struct {
	Volume dtyp.GUID
}

// GUID returns the VolumeID as a windows/guid.GUID for display and comparison.
func (v CVolumeId) GUID() guid.GUID { return v.Volume.GUID() }
