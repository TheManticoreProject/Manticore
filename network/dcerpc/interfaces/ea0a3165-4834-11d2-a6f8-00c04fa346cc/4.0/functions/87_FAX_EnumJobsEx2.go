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

// fAX_EnumJobsEx2Request carries the [in] parameters of FAX_EnumJobsEx2.
type fAX_EnumJobsEx2Request struct {
	FAllAccounts       ndr.BOOL
	LpcwstrAccountName *ndr.WSTR `ndr:"unique"`
	DwJobTypes         ndr.DWORD
	Level              ndr.DWORD
}

func (*fAX_EnumJobsEx2Request) Opnum() uint16 { return fax.OpnumFAX_EnumJobsEx2 }

// fAX_EnumJobsEx2Response carries the [out] parameters and return value of FAX_EnumJobsEx2.
type fAX_EnumJobsEx2Response struct {
	Buffer     []byte `ndr:"unique,conformant"`
	BufferSize ndr.DWORD
	LpdwJobs   ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// FAX_EnumJobsEx2 calls FAX_EnumJobsEx2 (opnum 87) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumJobsEx2(rpc ndr.Invoker, fAllAccounts ndr.BOOL, lpcwstrAccountName *ndr.WSTR, dwJobTypes ndr.DWORD, level ndr.DWORD) (Buffer []byte, BufferSize ndr.DWORD, LpdwJobs ndr.DWORD, err error) {
	req := &fAX_EnumJobsEx2Request{
		FAllAccounts:       fAllAccounts,
		LpcwstrAccountName: lpcwstrAccountName,
		DwJobTypes:         dwJobTypes,
		Level:              level,
	}
	var resp fAX_EnumJobsEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumJobsEx2: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	LpdwJobs = resp.LpdwJobs
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumJobsEx2 failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
