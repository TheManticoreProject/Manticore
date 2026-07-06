package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// s_DSDeleteObjectRequest carries the [in] parameters of S_DSDeleteObject.
type s_DSDeleteObjectRequest struct {
	DwObjectType ndr.DWORD
	PwcsPathName ndr.WSTR
}

func (*s_DSDeleteObjectRequest) Opnum() uint16 { return dscomm.OpnumS_DSDeleteObject }

// s_DSDeleteObjectResponse carries the [out] parameters and return value of S_DSDeleteObject.
type s_DSDeleteObjectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSDeleteObject calls S_DSDeleteObject (opnum 1) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSDeleteObject(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName ndr.WSTR) (err error) {
	req := &s_DSDeleteObjectRequest{
		DwObjectType: dwObjectType,
		PwcsPathName: pwcsPathName,
	}
	var resp s_DSDeleteObjectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSDeleteObject: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSDeleteObject failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
