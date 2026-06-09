package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegQueryMultipleValuesRequest carries the [in] parameters of BaseRegQueryMultipleValues.
type baseRegQueryMultipleValuesRequest struct {
	HKey       structures.RPC_HKEY
	Val_listIn []structures.RVALENT `ndr:"ref,size_is=Num_vals,varying,length_is=Num_vals"`
	Num_vals   ndr.DWORD
	LpvalueBuf []byte `ndr:"unique,varying"`
	LdwTotsize ndr.DWORD
}

func (*baseRegQueryMultipleValuesRequest) Opnum() uint16 {
	return winreg.OpnumBaseRegQueryMultipleValues
}

// baseRegQueryMultipleValuesResponse carries the [out] parameters and return value of BaseRegQueryMultipleValues.
type baseRegQueryMultipleValuesResponse struct {
	Val_listOut []structures.RVALENT `ndr:"ref,size_is=Num_vals,varying,length_is=Num_vals"`
	LpvalueBuf  []byte               `ndr:"unique,varying"`
	LdwTotsize  ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// BaseRegQueryMultipleValues calls BaseRegQueryMultipleValues (opnum 29) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegQueryMultipleValues(rpc ndr.Invoker, hKey structures.RPC_HKEY, val_listIn []structures.RVALENT, num_vals ndr.DWORD, lpvalueBuf []byte, ldwTotsize ndr.DWORD) (Val_listOut []structures.RVALENT, LpvalueBuf []byte, LdwTotsize ndr.DWORD, err error) {
	req := &baseRegQueryMultipleValuesRequest{
		HKey:       hKey,
		Val_listIn: val_listIn,
		Num_vals:   num_vals,
		LpvalueBuf: lpvalueBuf,
		LdwTotsize: ldwTotsize,
	}
	var resp baseRegQueryMultipleValuesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegQueryMultipleValues: %w", err)
		return
	}
	Val_listOut = resp.Val_listOut
	LpvalueBuf = resp.LpvalueBuf
	LdwTotsize = resp.LdwTotsize
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegQueryMultipleValues failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
