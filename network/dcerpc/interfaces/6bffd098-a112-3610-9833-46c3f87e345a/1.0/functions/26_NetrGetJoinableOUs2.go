package functions

// IDL source: [MS-WKST] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wkst/9fdbc753-0397-4236-bbfc-a380f9d23789
// A fetched copy is kept at ms-wkst.idl in the interface directory.

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrGetJoinableOUs2Request carries the [in] parameters of NetrGetJoinableOUs2.
type netrGetJoinableOUs2Request struct {
	ServerName      *ndr.WSTR `ndr:"unique"`
	DomainNameParam ndr.WSTR
	AccountName     *ndr.WSTR                              `ndr:"unique"`
	Password        *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	OUCount         ndr.DWORD
}

func (*netrGetJoinableOUs2Request) Opnum() uint16 { return wkssvc.OpnumNetrGetJoinableOUs2 }

// netrGetJoinableOUs2Response carries the [out] parameters and return value of NetrGetJoinableOUs2.
// OUs is [out,string,size_is(,*OUCount)] wchar_t***: the [ref] out pointer collapses, leaving a
// [unique] pointer to a conformant array of OUCount [unique] wide-string pointers (elem=unique).
type netrGetJoinableOUs2Response struct {
	OUCount ndr.DWORD
	OUs     []*ndr.WSTR `ndr:"unique,size_is=OUCount,elem=unique"`
	Status  ndr.DWORD   `ndr:"retval"`
}

// NetrGetJoinableOUs2 calls NetrGetJoinableOUs2 (opnum 26) ([MS-WKST] 3.2.4.4). It returns the
// organizational units into which the machine account could be created. OUCount is [in,out].
func NetrGetJoinableOUs2(rpc ndr.Invoker, serverName *ndr.WSTR, domainNameParam ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, oUCount ndr.DWORD) (OUCount ndr.DWORD, OUs []*ndr.WSTR, err error) {
	req := &netrGetJoinableOUs2Request{
		ServerName:      serverName,
		DomainNameParam: domainNameParam,
		AccountName:     accountName,
		Password:        password,
		OUCount:         oUCount,
	}
	var resp netrGetJoinableOUs2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrGetJoinableOUs2: %w", err)
		return
	}
	OUCount = resp.OUCount
	OUs = resp.OUs
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrGetJoinableOUs2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
