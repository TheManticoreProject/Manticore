package functions

// IDL source: [MS-RAIW] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-raiw/e59461f5-5486-4ec3-9ad6-14b784c1ecd6
// A fetched copy is kept at ms-raiw.idl in the interface directory.

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsGetDbRecsByNameRequest carries the [in] parameters of R_WinsGetDbRecsByName.
type r_WinsGetDbRecsByNameRequest struct {
	PWinsAdd *msraiw.WINSINTF_ADD_T `ndr:"unique"`
	Location ndr.DWORD
	// PName is [in, unique, size_is(NameLen + 1)] LPBYTE: a [unique] pointer to a
	// conformant byte buffer holding the NUL-terminated name. The size_is bound is an
	// arithmetic expression the codec cannot key on, so the maximum_count derives from
	// the slice length — callers pass a slice of NameLen+1 bytes (name plus a trailing
	// NUL) and set NameLen accordingly.
	PName           []uint8 `ndr:"unique,conformant"`
	NameLen         ndr.DWORD
	NoOfRecsDesired ndr.DWORD
	FOnlyStatic     ndr.DWORD
}

func (*r_WinsGetDbRecsByNameRequest) Opnum() uint16 { return winsif.OpnumR_WinsGetDbRecsByName }

// r_WinsGetDbRecsByNameResponse carries the [out] parameters and return value of R_WinsGetDbRecsByName.
type r_WinsGetDbRecsByNameResponse struct {
	PRecs  msraiw.WINSINTF_RECS_T
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsGetDbRecsByName calls R_WinsGetDbRecsByName (opnum 18) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsGetDbRecsByName(rpc ndr.Invoker, pWinsAdd *msraiw.WINSINTF_ADD_T, location ndr.DWORD, pName []uint8, nameLen ndr.DWORD, noOfRecsDesired ndr.DWORD, fOnlyStatic ndr.DWORD) (PRecs msraiw.WINSINTF_RECS_T, err error) {
	req := &r_WinsGetDbRecsByNameRequest{
		PWinsAdd:        pWinsAdd,
		Location:        location,
		PName:           pName,
		NameLen:         nameLen,
		NoOfRecsDesired: noOfRecsDesired,
		FOnlyStatic:     fOnlyStatic,
	}
	var resp r_WinsGetDbRecsByNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsGetDbRecsByName: %w", err)
		return
	}
	PRecs = resp.PRecs
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsGetDbRecsByName failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
