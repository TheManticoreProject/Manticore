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

// nspiQueryColumnsRequest carries the [in] parameters of NspiQueryColumns.
type nspiQueryColumnsRequest struct {
	HRpc     msnspi.NSPI_HANDLE
	Reserved ndr.DWORD
	DwFlags  ndr.DWORD
}

func (*nspiQueryColumnsRequest) Opnum() uint16 { return nspi.OpnumNspiQueryColumns }

// nspiQueryColumnsResponse carries the [out] parameters and return value of NspiQueryColumns.
type nspiQueryColumnsResponse struct {
	PpColumns *msnspi.PropertyTagArray_r `ndr:"unique"`
	Status    ndr.DWORD                  `ndr:"retval"`
}

// NspiQueryColumns calls NspiQueryColumns (opnum 16) ([MS-NSPI] — verify the parameter
// modeling and status handling).
func NspiQueryColumns(rpc ndr.Invoker, hRpc msnspi.NSPI_HANDLE, reserved ndr.DWORD, dwFlags ndr.DWORD) (PpColumns *msnspi.PropertyTagArray_r, err error) {
	req := &nspiQueryColumnsRequest{
		HRpc:     hRpc,
		Reserved: reserved,
		DwFlags:  dwFlags,
	}
	var resp nspiQueryColumnsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NspiQueryColumns: %w", err)
		return
	}
	PpColumns = resp.PpColumns
	if uint32(resp.Status) != nspi.StatusSuccess {
		err = fmt.Errorf("NspiQueryColumns failed: %s", nspi.StatusString(uint32(resp.Status)))
	}
	return
}
