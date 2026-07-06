package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rasRpcDeleteEntryRequest carries the [in] parameters of RasRpcDeleteEntry.
type rasRpcDeleteEntryRequest struct {
	LpszPhonebook ndr.WSTR
	LpszEntry     ndr.WSTR
}

func (*rasRpcDeleteEntryRequest) Opnum() uint16 { return rasrpc.OpnumRasRpcDeleteEntry }

// rasRpcDeleteEntryResponse carries the [out] parameters and return value of RasRpcDeleteEntry.
type rasRpcDeleteEntryResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RasRpcDeleteEntry calls RasRpcDeleteEntry (opnum 5) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcDeleteEntry(rpc ndr.Invoker, lpszPhonebook ndr.WSTR, lpszEntry ndr.WSTR) (err error) {
	req := &rasRpcDeleteEntryRequest{
		LpszPhonebook: lpszPhonebook,
		LpszEntry:     lpszEntry,
	}
	var resp rasRpcDeleteEntryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcDeleteEntry: %w", err)
		return
	}
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcDeleteEntry failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
