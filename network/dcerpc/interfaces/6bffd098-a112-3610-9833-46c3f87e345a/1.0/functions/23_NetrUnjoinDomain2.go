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

// netrUnjoinDomain2Request carries the [in] parameters of NetrUnjoinDomain2.
type netrUnjoinDomain2Request struct {
	ServerName  *ndr.WSTR                              `ndr:"unique"`
	AccountName *ndr.WSTR                              `ndr:"unique"`
	Password    *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	Options     ndr.DWORD
}

func (*netrUnjoinDomain2Request) Opnum() uint16 { return wkssvc.OpnumNetrUnjoinDomain2 }

// netrUnjoinDomain2Response carries the [out] parameters and return value of NetrUnjoinDomain2.
type netrUnjoinDomain2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrUnjoinDomain2 calls NetrUnjoinDomain2 (opnum 23) ([MS-WKST] 3.2.4).
func NetrUnjoinDomain2(rpc ndr.Invoker, serverName *ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, options ndr.DWORD) (err error) {
	req := &netrUnjoinDomain2Request{
		ServerName:  serverName,
		AccountName: accountName,
		Password:    password,
		Options:     options,
	}
	var resp netrUnjoinDomain2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrUnjoinDomain2: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrUnjoinDomain2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
