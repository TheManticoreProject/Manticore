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

// baseRegQueryMultipleValues2Request carries the [in] parameters of BaseRegQueryMultipleValues2.
type baseRegQueryMultipleValues2Request struct {
	HKey       msrrp.RPC_HKEY
	Val_listIn []msrrp.RVALENT `ndr:"ref,size_is=Num_vals,varying,length_is=Num_vals"`
	Num_vals   ndr.DWORD
	LpvalueBuf []byte `ndr:"unique,varying"`
	LdwTotsize ndr.DWORD
}

func (*baseRegQueryMultipleValues2Request) Opnum() uint16 {
	return winreg.OpnumBaseRegQueryMultipleValues2
}

// baseRegQueryMultipleValues2Response carries the [out] parameters and return value of BaseRegQueryMultipleValues2.
type baseRegQueryMultipleValues2Response struct {
	Val_listOut     []msrrp.RVALENT `ndr:"ref,size_is=Num_vals,varying,length_is=Num_vals"`
	LpvalueBuf      []byte          `ndr:"unique,varying"`
	LdwRequiredSize ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// BaseRegQueryMultipleValues2 calls BaseRegQueryMultipleValues2 (opnum 34) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegQueryMultipleValues2(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, val_listIn []msrrp.RVALENT, num_vals ndr.DWORD, lpvalueBuf []byte, ldwTotsize ndr.DWORD) (Val_listOut []msrrp.RVALENT, LpvalueBuf []byte, LdwRequiredSize ndr.DWORD, err error) {
	req := &baseRegQueryMultipleValues2Request{
		HKey:       hKey,
		Val_listIn: val_listIn,
		Num_vals:   num_vals,
		LpvalueBuf: lpvalueBuf,
		LdwTotsize: ldwTotsize,
	}
	var resp baseRegQueryMultipleValues2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegQueryMultipleValues2: %w", err)
		return
	}
	Val_listOut = resp.Val_listOut
	LpvalueBuf = resp.LpvalueBuf
	LdwRequiredSize = resp.LdwRequiredSize
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegQueryMultipleValues2 failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
