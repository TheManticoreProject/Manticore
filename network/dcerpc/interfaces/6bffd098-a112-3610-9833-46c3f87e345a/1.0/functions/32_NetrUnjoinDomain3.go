package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrUnjoinDomain3Request carries the [in] parameters of NetrUnjoinDomain3.
type netrUnjoinDomain3Request struct {
	ServerName  *ndr.WSTR                                  `ndr:"unique"`
	AccountName *ndr.WSTR                                  `ndr:"unique"`
	Password    *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES `ndr:"unique"`
	Options     ndr.DWORD
}

func (*netrUnjoinDomain3Request) Opnum() uint16 { return wkssvc.OpnumNetrUnjoinDomain3 }

// netrUnjoinDomain3Response carries the [out] parameters and return value of NetrUnjoinDomain3.
type netrUnjoinDomain3Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrUnjoinDomain3 calls NetrUnjoinDomain3 (opnum 32) ([MS-WKST] 3.2.4).
func NetrUnjoinDomain3(rpc ndr.Invoker, serverName *ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES, options ndr.DWORD) (err error) {
	req := &netrUnjoinDomain3Request{
		ServerName:  serverName,
		AccountName: accountName,
		Password:    password,
		Options:     options,
	}
	var resp netrUnjoinDomain3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrUnjoinDomain3: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrUnjoinDomain3 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
