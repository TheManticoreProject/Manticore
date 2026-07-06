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

// netrJoinDomain3Request carries the [in] parameters of NetrJoinDomain3.
type netrJoinDomain3Request struct {
	ServerName       *ndr.WSTR `ndr:"unique"`
	DomainNameParam  ndr.WSTR
	MachineAccountOU *ndr.WSTR                                  `ndr:"unique"`
	AccountName      *ndr.WSTR                                  `ndr:"unique"`
	Password         *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES `ndr:"unique"`
	Options          ndr.DWORD
}

func (*netrJoinDomain3Request) Opnum() uint16 { return wkssvc.OpnumNetrJoinDomain3 }

// netrJoinDomain3Response carries the [out] parameters and return value of NetrJoinDomain3.
type netrJoinDomain3Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrJoinDomain3 calls NetrJoinDomain3 (opnum 31) ([MS-WKST] 3.2.4).
func NetrJoinDomain3(rpc ndr.Invoker, serverName *ndr.WSTR, domainNameParam ndr.WSTR, machineAccountOU *ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES, options ndr.DWORD) (err error) {
	req := &netrJoinDomain3Request{
		ServerName:       serverName,
		DomainNameParam:  domainNameParam,
		MachineAccountOU: machineAccountOU,
		AccountName:      accountName,
		Password:         password,
		Options:          options,
	}
	var resp netrJoinDomain3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrJoinDomain3: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrJoinDomain3 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
