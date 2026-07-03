package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// Restriction_r ([MS-NSPI] 2.2.5.3). Res is a non-encapsulated RestrictionUnion_r whose
// discriminant is rt; the discriminant is transmitted inline, so it must be set before
// marshalling (see SetDiscriminants).
type Restriction_r struct {
	Rt  ndr.DWORD
	Res RestrictionUnion_r
}

// Restriction type tags ([MS-NSPI] 2.2.5.3 — the rt / RestrictionUnion_r discriminant).
const (
	RestrictionTypeAnd          ndr.DWORD = 0x00000000
	RestrictionTypeOr           ndr.DWORD = 0x00000001
	RestrictionTypeNot          ndr.DWORD = 0x00000002
	RestrictionTypeContent      ndr.DWORD = 0x00000003
	RestrictionTypeProperty     ndr.DWORD = 0x00000004
	RestrictionTypeCompareProps ndr.DWORD = 0x00000005
	RestrictionTypeBitMask      ndr.DWORD = 0x00000006
	RestrictionTypeSize         ndr.DWORD = 0x00000007
	RestrictionTypeExist        ndr.DWORD = 0x00000008
	RestrictionTypeSubObject    ndr.DWORD = 0x00000009
)

// SetDiscriminants derives Res.Tag from Rt and recurses into every sub-restriction and
// embedded PropertyValue_r, so a caller that populated only the selected arm and its rt
// produces a wire-correct (fully discriminated) restriction tree. It operates in place; the
// values written are the discriminants already implied by rt/ulPropTag.
func (r *Restriction_r) SetDiscriminants() {
	if r == nil {
		return
	}
	r.Res.Tag = int32(r.Rt)
	switch r.Rt {
	case RestrictionTypeAnd:
		for i := range r.Res.ResAnd.LpRes {
			r.Res.ResAnd.LpRes[i].SetDiscriminants()
		}
	case RestrictionTypeOr:
		for i := range r.Res.ResOr.LpRes {
			r.Res.ResOr.LpRes[i].SetDiscriminants()
		}
	case RestrictionTypeNot:
		r.Res.ResNot.LpRes.SetDiscriminants()
	case RestrictionTypeContent:
		r.Res.ResContent.LpProp.SetDiscriminant()
	case RestrictionTypeProperty:
		r.Res.ResProperty.LpProp.SetDiscriminant()
	case RestrictionTypeSubObject:
		r.Res.ResSubRestriction.LpRes.SetDiscriminants()
	}
}
