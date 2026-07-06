package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrValidateName3Request carries the [in] parameters of NetrValidateName3.
type netrValidateName3Request struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	NameToValidate ndr.WSTR
	AccountName    *ndr.WSTR                                  `ndr:"unique"`
	Password       *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES `ndr:"unique"`
	NameType       mswkst.NETSETUP_NAME_TYPE
}

func (*netrValidateName3Request) Opnum() uint16 { return wkssvc.OpnumNetrValidateName3 }

// netrValidateName3Response carries the [out] parameters and return value of NetrValidateName3.
type netrValidateName3Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrValidateName3 calls NetrValidateName3 (opnum 34) ([MS-WKST] 3.2.4).
func NetrValidateName3(rpc ndr.Invoker, serverName *ndr.WSTR, nameToValidate ndr.WSTR, accountName *ndr.WSTR, password *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES, nameType mswkst.NETSETUP_NAME_TYPE) (err error) {
	req := &netrValidateName3Request{
		ServerName:     serverName,
		NameToValidate: nameToValidate,
		AccountName:    accountName,
		Password:       password,
		NameType:       nameType,
	}
	var resp netrValidateName3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrValidateName3: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrValidateName3 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
