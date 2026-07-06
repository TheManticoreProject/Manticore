package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrRemoveAlternateComputerName2Request carries the [in] parameters of NetrRemoveAlternateComputerName2.
type netrRemoveAlternateComputerName2Request struct {
	ServerName        *ndr.WSTR                                  `ndr:"unique"`
	AlternateName     *ndr.WSTR                                  `ndr:"unique"`
	DomainAccount     *ndr.WSTR                                  `ndr:"unique"`
	EncryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES `ndr:"unique"`
	Reserved          ndr.DWORD
}

func (*netrRemoveAlternateComputerName2Request) Opnum() uint16 {
	return wkssvc.OpnumNetrRemoveAlternateComputerName2
}

// netrRemoveAlternateComputerName2Response carries the [out] parameters and return value of NetrRemoveAlternateComputerName2.
type netrRemoveAlternateComputerName2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrRemoveAlternateComputerName2 calls NetrRemoveAlternateComputerName2 (opnum 36) ([MS-WKST] 3.2.4).
func NetrRemoveAlternateComputerName2(rpc ndr.Invoker, serverName *ndr.WSTR, alternateName *ndr.WSTR, domainAccount *ndr.WSTR, encryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES, reserved ndr.DWORD) (err error) {
	req := &netrRemoveAlternateComputerName2Request{
		ServerName:        serverName,
		AlternateName:     alternateName,
		DomainAccount:     domainAccount,
		EncryptedPassword: encryptedPassword,
		Reserved:          reserved,
	}
	var resp netrRemoveAlternateComputerName2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrRemoveAlternateComputerName2: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrRemoveAlternateComputerName2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
