package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_AccessCheckEx2Request carries the [in] parameters of FAX_AccessCheckEx2.
type fAX_AccessCheckEx2Request struct {
	AccessMask ndr.DWORD
	LpdwRights *ndr.DWORD `ndr:"unique"`
}

func (*fAX_AccessCheckEx2Request) Opnum() uint16 { return fax.OpnumFAX_AccessCheckEx2 }

// fAX_AccessCheckEx2Response carries the [out] parameters and return value of FAX_AccessCheckEx2.
type fAX_AccessCheckEx2Response struct {
	PfAccess   ndr.BOOL
	LpdwRights *ndr.DWORD `ndr:"unique"`
	Status     ndr.DWORD  `ndr:"retval"`
}

// FAX_AccessCheckEx2 calls FAX_AccessCheckEx2 (opnum 100) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_AccessCheckEx2(rpc ndr.Invoker, accessMask ndr.DWORD, lpdwRights *ndr.DWORD) (PfAccess ndr.BOOL, LpdwRights *ndr.DWORD, err error) {
	req := &fAX_AccessCheckEx2Request{
		AccessMask: accessMask,
		LpdwRights: lpdwRights,
	}
	var resp fAX_AccessCheckEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_AccessCheckEx2: %w", err)
		return
	}
	PfAccess = resp.PfAccess
	LpdwRights = resp.LpdwRights
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_AccessCheckEx2 failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
