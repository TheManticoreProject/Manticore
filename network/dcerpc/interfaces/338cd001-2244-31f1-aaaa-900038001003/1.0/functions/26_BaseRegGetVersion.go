package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegGetVersionRequest carries the [in] parameters of BaseRegGetVersion.
type baseRegGetVersionRequest struct {
	HKey msrrp.RPC_HKEY
}

func (*baseRegGetVersionRequest) Opnum() uint16 { return winreg.OpnumBaseRegGetVersion }

// baseRegGetVersionResponse carries the [out] parameters and return value of BaseRegGetVersion.
type baseRegGetVersionResponse struct {
	LpdwVersion ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// BaseRegGetVersion calls BaseRegGetVersion (opnum 26) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegGetVersion(rpc ndr.Invoker, hKey msrrp.RPC_HKEY) (LpdwVersion ndr.DWORD, err error) {
	req := &baseRegGetVersionRequest{
		HKey: hKey,
	}
	var resp baseRegGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegGetVersion: %w", err)
		return
	}
	LpdwVersion = resp.LpdwVersion
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegGetVersion failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
