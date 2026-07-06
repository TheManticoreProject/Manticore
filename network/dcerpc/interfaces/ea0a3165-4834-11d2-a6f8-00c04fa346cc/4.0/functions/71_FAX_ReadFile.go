package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_ReadFileRequest carries the [in] parameters of FAX_ReadFile.
type fAX_ReadFileRequest struct {
	HCopy         msfax.RPC_FAX_COPY_HANDLE
	DwMaxDataSize ndr.DWORD
	LpdwDataSize  ndr.DWORD
}

func (*fAX_ReadFileRequest) Opnum() uint16 { return fax.OpnumFAX_ReadFile }

// fAX_ReadFileResponse carries the [out] parameters and return value of FAX_ReadFile.
type fAX_ReadFileResponse struct {
	LpbData      []uint8 `ndr:"ref,conformant"`
	LpdwDataSize ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// FAX_ReadFile calls FAX_ReadFile (opnum 71) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_ReadFile(rpc ndr.Invoker, hCopy msfax.RPC_FAX_COPY_HANDLE, dwMaxDataSize ndr.DWORD, lpdwDataSize ndr.DWORD) (LpbData []uint8, LpdwDataSize ndr.DWORD, err error) {
	req := &fAX_ReadFileRequest{
		HCopy:         hCopy,
		DwMaxDataSize: dwMaxDataSize,
		LpdwDataSize:  lpdwDataSize,
	}
	var resp fAX_ReadFileResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_ReadFile: %w", err)
		return
	}
	LpbData = resp.LpbData
	LpdwDataSize = resp.LpdwDataSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_ReadFile failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
