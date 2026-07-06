package functions

// IDL source: [MS-W32T] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-w32t/5793e908-22ef-4cea-962f-fca8a72c485a
// A fetched copy is kept at ms-w32t.idl in the interface directory.

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// w32TimeGetNetlogonServiceBitsRequest carries the [in] parameters of W32TimeGetNetlogonServiceBits.
type w32TimeGetNetlogonServiceBitsRequest struct {
}

func (*w32TimeGetNetlogonServiceBitsRequest) Opnum() uint16 {
	return W32Time.OpnumW32TimeGetNetlogonServiceBits
}

// w32TimeGetNetlogonServiceBitsResponse carries the [out] parameters and return value of W32TimeGetNetlogonServiceBits.
type w32TimeGetNetlogonServiceBitsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// W32TimeGetNetlogonServiceBits calls W32TimeGetNetlogonServiceBits (opnum 1) ([MS-W32T] section 3.2.4).
func W32TimeGetNetlogonServiceBits(rpc ndr.Invoker) (err error) {
	req := &w32TimeGetNetlogonServiceBitsRequest{}
	var resp w32TimeGetNetlogonServiceBitsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeGetNetlogonServiceBits: %w", err)
		return
	}
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeGetNetlogonServiceBits failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}
