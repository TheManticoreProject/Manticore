package functions

// IDL source: [MS-NSPI] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nspi/2554418c-a060-473a-950a-e009a53e33d9
// A fetched copy is kept at ms-nspi.idl in the interface directory.

import (
	"fmt"

	nspi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc5a18-4264-101a-8c59-08002b2f8426/56.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnspi "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nspi"
)

// nspiResolveNamesRequest carries the [in] parameters of NspiResolveNames.
type nspiResolveNamesRequest struct {
	HRpc      msnspi.NSPI_HANDLE
	Reserved  ndr.DWORD
	PStat     msnspi.STAT
	PPropTags *msnspi.PropertyTagArray_r `ndr:"unique"`
	PaStr     msnspi.StringsArray_r
}

func (*nspiResolveNamesRequest) Opnum() uint16 { return nspi.OpnumNspiResolveNames }

// nspiResolveNamesResponse carries the [out] parameters and return value of NspiResolveNames.
type nspiResolveNamesResponse struct {
	PpMIds *msnspi.PropertyTagArray_r `ndr:"unique"`
	PpRows *msnspi.PropertyRowSet_r   `ndr:"unique"`
	Status ndr.DWORD                  `ndr:"retval"`
}

// NspiResolveNames calls NspiResolveNames (opnum 19) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiResolveNames(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, pStat msnspi.STAT, pPropTags *msnspi.PropertyTagArray_r, paStr msnspi.StringsArray_r) (PpMIds *msnspi.PropertyTagArray_r, PpRows *msnspi.PropertyRowSet_r, err error) {
	req := &nspiResolveNamesRequest{
		HRpc:      hRpc,
		Reserved:  reserved,
		PStat:     pStat,
		PPropTags: pPropTags,
		PaStr:     paStr,
	}
	var resp nspiResolveNamesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiResolveNames: %w", err)
		return
	}
	PpMIds = resp.PpMIds
	PpRows = resp.PpRows
	// ErrorsReturned is a success-severity warning: the call succeeded overall but one or
	// more returned properties are PtypErrorCode ([MS-NSPI] 3.1.4.1.19 / [MS-OXCDATA] 2.5).
	if s := uint32(resp.Status); s != nspi.StatusSuccess && s != nspi.StatusErrorsReturned {
		err = fmt.Errorf("NspiResolveNames failed: %s", nspi.StatusString(s))
	}
	return
}
